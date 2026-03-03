import { spawn, execSync } from "child_process"
import { readFileSync, unlinkSync, existsSync } from "fs"
import { join } from "path"
import { tmpdir, homedir } from "os"

const DEBOUNCE_MS = 10000

const configPath = join(
  process.env.XDG_CONFIG_HOME || join(homedir(), ".config"),
  "ghost-tab",
  "opencode-features.json"
)
let features = { sound: false, spinner: false }
try {
  features = JSON.parse(readFileSync(configPath, "utf-8"))
} catch {}

function getProject(): string {
  try {
    const session = execSync("tmux display-message -p '#S'", {
      stdio: ["pipe", "pipe", "ignore"],
    }).toString().trim()
    if (session) return session
  } catch {}
  return process.cwd().split("/").pop() || "opencode"
}

function getPidFile(): string {
  return join(tmpdir(), `ghost-tab-spinner-${getProject()}.pid`)
}

function killSpinner(): void {
  const pf = getPidFile()
  if (existsSync(pf)) {
    try {
      const pid = parseInt(readFileSync(pf, "utf-8").trim())
      process.kill(pid)
    } catch {}
    try { unlinkSync(pf) } catch {}
    const project = getProject()
    process.stdout.write(`\x1b]0;${project}\x07`)
  }
}

function startSpinner(): void {
  const project = getProject()
  const frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
  const script = `
    echo $$ > "${getPidFile()}"
    FRAMES=(${frames.join(" ")})
    i=0
    while true; do
      printf '\\033]0;%s %s\\007' "\${FRAMES[\$i]}" "${project}"
      i=$(( (i + 1) % \${#FRAMES[@]} ))
      sleep 0.1
    done
  `
  const proc = spawn("bash", ["-c", script], { stdio: "ignore", detached: true })
  proc.unref()
}

// Debounce timer: delays sound and spinner until the AI is genuinely idle.
// Subagent completions cause brief idle events (2-15+ seconds of thinking
// before the next action), so we wait DEBOUNCE_MS before notifying.
let idleTimer: ReturnType<typeof setTimeout> | null = null

function cancelIdleTimer(): void {
  if (idleTimer) {
    clearTimeout(idleTimer)
    idleTimer = null
  }
}

function onIdle(): void {
  if (features.sound) {
    spawn("afplay", ["/System/Library/Sounds/Bottle.aiff"], { stdio: "ignore" })
  }
  if (features.spinner) {
    killSpinner()
    startSpinner()
  }
  idleTimer = null
}

export const GhostTab = async () => {
  return {
    event: async ({ event }: { event: { type: string; properties?: any } }) => {
      if (event.type === "session.idle") {
        // Don't fire immediately — debounce to filter out subagent completions.
        cancelIdleTimer()
        idleTimer = setTimeout(onIdle, DEBOUNCE_MS)
      }
      if (event.type === "session.status") {
        const status = event.properties?.status
        if (status?.type === "busy") {
          // AI started working — cancel pending notification and stop spinner.
          cancelIdleTimer()
          if (features.spinner) {
            killSpinner()
          }
        }
      }
    },
  }
}
