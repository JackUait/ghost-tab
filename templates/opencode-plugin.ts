import { spawn, execSync } from "child_process"
import { readFileSync, unlinkSync, existsSync } from "fs"
import { join } from "path"
import { tmpdir, homedir } from "os"

const SHORT_DEBOUNCE_MS = 2000
const LONG_DEBOUNCE_MS = 15000
const COOLDOWN_WINDOW_MS = 30000

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

// Track last tool completion for cooldown-based debounce.
// When a tool completed recently (within COOLDOWN_WINDOW_MS), use the long
// debounce to filter out subagent processing gaps. Otherwise, use the short
// debounce for fast genuine-idle notification.
let lastToolCompleteTime = 0

// Debounce timer: delays sound and spinner until the AI is genuinely idle.
let idleTimer: ReturnType<typeof setTimeout> | null = null

function cancelIdleTimer(): void {
  if (idleTimer) {
    clearTimeout(idleTimer)
    idleTimer = null
  }
}

function getDebounceMs(): number {
  return (Date.now() - lastToolCompleteTime < COOLDOWN_WINDOW_MS)
    ? LONG_DEBOUNCE_MS
    : SHORT_DEBOUNCE_MS
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
    "tool.execute.after": async () => {
      lastToolCompleteTime = Date.now()
    },
    event: async ({ event }: { event: { type: string; properties?: any } }) => {
      if (event.type === "session.idle") {
        // Debounce with cooldown-aware threshold: short when no recent tool
        // activity, long after tool use to filter subagent processing gaps.
        cancelIdleTimer()
        idleTimer = setTimeout(onIdle, getDebounceMs())
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
