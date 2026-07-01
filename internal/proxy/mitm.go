package proxy

// MITM forward-proxy support: a local CA + leaf cert lifecycle and a transparent
// CONNECT relay, mirroring teamclaude's src/mitm.js. When claude is launched with
// HTTPS_PROXY pointed at us it sends `CONNECT api.anthropic.com:443`; we present
// a locally-minted leaf (trusted by the client via NODE_EXTRA_CA_CERTS — a
// per-process trust, never the system keychain), decrypt the stream, inject the
// active account's token, and forward upstream. A host routing table decides
// per-CONNECT behavior: the upstream host is rewritten, a built-in test host is
// answered locally, and anything else is blind-tunneled.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TestHost is a host the MITM proxy always intercepts and answers itself, so the
// proxy + CA can be verified end-to-end with no credentials, e.g.:
//
//	curl --proxy http://localhost:PORT --cacert <ca.pem> https://www.example.org/
const TestHost = "www.example.org"

const (
	caCertFile  = "wisp-deck-ca.pem"
	caKeyFile   = "wisp-deck-ca.key"
	leafCertF   = "wisp-deck-leaf.pem"
	leafKeyFile = "wisp-deck-leaf.key"
)

// Certs is the result of EnsureCerts: the on-disk CA cert path (for
// NODE_EXTRA_CA_CERTS) and the leaf certificate to present to clients.
type Certs struct {
	CACertPath string
	Leaf       tls.Certificate
}

// EnsureCerts loads a cached CA + leaf covering host (and TestHost) from dir, or
// generates and persists them when missing/invalid. The CA cert is world-
// readable (clients trust it via NODE_EXTRA_CA_CERTS); private keys are 0600.
func EnsureCerts(dir, host string) (Certs, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Certs{}, err
	}
	hosts := []string{host, TestHost}
	if host == TestHost {
		hosts = []string{TestHost}
	}

	// Try the cached chain first.
	if leaf, err := tls.LoadX509KeyPair(filepath.Join(dir, leafCertF), filepath.Join(dir, leafKeyFile)); err == nil {
		if caPEM, err := os.ReadFile(filepath.Join(dir, caCertFile)); err == nil && leafCovers(caPEM, leaf, hosts) {
			return Certs{CACertPath: filepath.Join(dir, caCertFile), Leaf: leaf}, nil
		}
	}

	caCertPEM, caKeyPEM, caCert, caKey, err := generateCA()
	if err != nil {
		return Certs{}, err
	}
	leafCertPEM, leafKeyPEM, err := generateLeaf(hosts, caCert, caKey)
	if err != nil {
		return Certs{}, err
	}
	if err := writeAtomic(filepath.Join(dir, caCertFile), caCertPEM, 0o644); err != nil {
		return Certs{}, err
	}
	if err := writeAtomic(filepath.Join(dir, caKeyFile), caKeyPEM, 0o600); err != nil {
		return Certs{}, err
	}
	if err := writeAtomic(filepath.Join(dir, leafCertF), leafCertPEM, 0o644); err != nil {
		return Certs{}, err
	}
	if err := writeAtomic(filepath.Join(dir, leafKeyFile), leafKeyPEM, 0o600); err != nil {
		return Certs{}, err
	}
	leaf, err := tls.X509KeyPair(leafCertPEM, leafKeyPEM)
	if err != nil {
		return Certs{}, err
	}
	return Certs{CACertPath: filepath.Join(dir, caCertFile), Leaf: leaf}, nil
}

// leafCovers reports whether leaf is signed by the CA in caPEM and valid for
// every host.
func leafCovers(caPEM []byte, leaf tls.Certificate, hosts []string) bool {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return false
	}
	x, err := x509.ParseCertificate(leaf.Certificate[0])
	if err != nil {
		return false
	}
	for _, h := range hosts {
		if _, err := x.Verify(x509.VerifyOptions{Roots: pool, DNSName: h}); err != nil {
			return false
		}
	}
	return true
}

func generateCA() (certPEM, keyPEM []byte, cert *x509.Certificate, key *ecdsa.PrivateKey, err error) {
	key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          bigSerial(),
		Subject:               pkix.Name{CommonName: "Wisp Deck Local CA", Organization: []string{"Wisp Deck"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cert, err = x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, cert, key, nil
}

func generateLeaf(hosts []string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (certPEM, keyPEM []byte, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: bigSerial(),
		Subject:      pkix.Name{CommonName: hosts[0]},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     hosts,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, nil
}

func bigSerial() *big.Int {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return n
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp := fmt.Sprintf("%s.tmp%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// --- CONNECT / MITM request handling ---------------------------------------

// hostMode decides per-CONNECT behavior: "rewrite" (MITM + token inject) for the
// upstream host, "test" for the built-in TestHost, "tunnel" (blind) otherwise.
func (s *Server) hostMode(host string) string {
	switch host {
	case TestHost:
		return "test"
	case s.mitmHost:
		return "rewrite"
	default:
		return "tunnel"
	}
}

// handleConnect implements the forward-proxy CONNECT: it hijacks the client
// connection, acknowledges the tunnel, then either MITM-intercepts, answers the
// test host locally, or blind-tunnels to the requested host.
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host, port = r.Host, "443"
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking unsupported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		return
	}
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		clientConn.Close()
		return
	}

	mode := s.hostMode(host)
	// Without a leaf we cannot terminate TLS, so fall back to a blind tunnel.
	if s.certs == nil && mode != "tunnel" {
		mode = "tunnel"
	}
	switch mode {
	case "rewrite":
		s.serveTunneledTLS(clientConn, http.HandlerFunc(func(rw http.ResponseWriter, rr *http.Request) {
			s.proxyRequest(rw, rr)
		}))
	case "test":
		s.serveTunneledTLS(clientConn, http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("wisp-deck MITM proxy OK\n"))
		}))
	default:
		blindTunnel(clientConn, net.JoinHostPort(host, port))
	}
}

// serveTunneledTLS terminates TLS on clientConn with the local leaf and serves
// the decrypted HTTP/1.1 stream with handler until the connection closes.
func (s *Server) serveTunneledTLS(clientConn net.Conn, handler http.Handler) {
	tlsConn := tls.Server(clientConn, s.tlsConf)
	if err := tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return
	}
	l := &oneConnListener{ch: make(chan net.Conn, 1), closed: make(chan struct{}), addr: tlsConn.LocalAddr()}
	l.ch <- &notifyConn{Conn: tlsConn, onClose: l.Close}
	srv := &http.Server{
		Handler:     handler,
		IdleTimeout: 60 * time.Second,
	}
	_ = srv.Serve(l) // returns once the single connection is closed
}

// blindTunnel copies bytes verbatim between the client and target (no TLS
// interception) — used for hosts we don't rewrite.
func blindTunnel(clientConn net.Conn, target string) {
	defer clientConn.Close()
	upstreamConn, err := net.DialTimeout("tcp", target, 15*time.Second)
	if err != nil {
		return
	}
	defer upstreamConn.Close()
	done := make(chan struct{}, 2)
	cp := func(dst, src net.Conn) { _, _ = io.Copy(dst, src); done <- struct{}{} }
	go cp(upstreamConn, clientConn)
	go cp(clientConn, upstreamConn)
	<-done
}

// oneConnListener is a net.Listener that yields exactly one pre-accepted
// connection, then blocks until closed — letting http.Server serve a single
// hijacked/TLS connection with full keep-alive support.
type oneConnListener struct {
	ch     chan net.Conn
	closed chan struct{}
	once   sync.Once
	addr   net.Addr
}

func (l *oneConnListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.ch:
		return c, nil
	case <-l.closed:
		return nil, errListenerClosed
	}
}

func (l *oneConnListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *oneConnListener) Addr() net.Addr { return l.addr }

var errListenerClosed = fmt.Errorf("wisp-deck-proxy: listener closed")

// notifyConn closes the owning one-shot listener when the connection closes, so
// http.Server.Serve returns once the tunnel ends.
type notifyConn struct {
	net.Conn
	onClose func() error
	once    sync.Once
}

func (c *notifyConn) Close() error {
	c.once.Do(func() { _ = c.onClose() })
	return c.Conn.Close()
}
