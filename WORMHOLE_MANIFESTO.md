# Wormhole: The Secure Bridge for Cloud Operations

## 🚀 Vision & Objective
**Wormhole** (`wh`) is a high-performance, interactive CLI tool designed to replace legacy Bash/Python scripts with a unified, professional, and "agent-friendly" platform. 

The goal is to provide a **premium operational experience** that abstracts the complexity of AWS infrastructure (tunnels, bridge instances, credential discovery) into simple, beautiful commands. It acts as a secure "wormhole" into private networks.

---

## 🏗 Architecture & Technology Stack
Wormhole is built with **Go 2026 stable**, following modern CLI design patterns:

*   **Command Framework**: `github.com/spf13/cobra` (invoked as `wh`).
*   **Interactive UX**: `github.com/charmbracelet/huh` (multi-step forms and selection menus).
*   **Aesthetics**: `github.com/charmbracelet/lipgloss` (rich terminal styling and colors).
*   **Cloud Logic**: Official `aws-sdk-go-v2` (high-performance, context-aware SDK).

---

## 🧠 Core Logic & Heuristics

Wormhole stands out by its ability to "guess" and "discover" connection paths without manual configuration.

### 1. Metadata Discovery Hierarchy
To connect to an RDS, Wormhole follows this strict order of discovery:
1.  **AWS Resource Tags** (State of the Art): Looks for `vops:credential-path` and `vops:bridge-cluster` on the RDS instance.
2.  **Manual Mapping**: Checks `internal/aws/mapping.go` for hardcoded overrides.
3.  **Enhanced Fuzzy Search**: Scans Parameter Store for paths. The algorithm penalizes partial matches and rewards specificity (e.g., distinguishing between `rds` and `web-rds` accurately).

### 2. Bridge-to-DB Connectivity & Robustness
Since RDS instances are in private subnets, Wormhole creates a secure bridge:
- **Dynamic Bridge Discovery**: Finds the most suitable ECS cluster based on naming heuristics or resource tags.
- **Port Randomization**: To avoid "zombie" tunnels from previous failed sessions, each connection uses a random local port (5440-5900).
- **Clean Exit**: System resources are automatically cleaned up on session termination.

---

## 🛠 Features implemented

### **ECS (Elastic Container Service)**
- `wh ecs ls`: Lists clusters with interactive filtering.
- `wh ecs conn`: Launches a secure session into a container using `aws ecs execute-command`.

### **RDS (Relational Database Service)**
- `wh rds ls`: Lists all RDS instances in the account.
- `wh rds conn`: Automates discovery, tunneling, and launches `psql` with a masked credential summary.
- `wh rds check-access`: **Diagnostic Tool**. Scans Security Group rules and lists all running EC2 instances that have network line-of-sight to the RDS.

---

## ✨ Premium Operational Experience
Wormhole is not just functional; it's beautiful.
- **Visual Summaries**: Every connection starts with a Lipgloss-styled table showing the exact parameters being used.
- **Real-time Feedback**: Use of emojis and status indicators (`🚀`, `⏳`, `🔌`) to keep the operator informed.
- **Masked Security**: Passwords and sensitive paths are partially masked in logs to allow safe screen sharing.

---

## 🔮 Future Vision: The Agent Era
Wormhole is designed to be evolved by AI agents. Future roadmap includes:

1.  **Advanced Diagnostic Engine**: Expanding `check-access` to automatically verify IAM policies, KMS permissions, and DB-level user roles.
2.  **TUI Dashboard**: Migrating from a command-line flow to a persistent TUI (Bubble Tea) dashboard for real-time monitoring.
3.  **Cross-Cloud Expansion**: Support for GCP/Azure following the same "Discovery & Tunnel" philosophy.
4.  **Automatic Credential Rotation**: Ability to rotate passwords in SSM/RDS directly from the CLI.

---

## 📝 Guidelines for Future Agents
- **Documentation**: Always include a concise docstring at the top of every new file.
- **Robustness**: Never assume a resource exists. Always implement fallback loops (like the bridge discovery).
- **UX First**: Use `lipgloss` and `huh` to maintain the premium aesthetic.
- **Clean Environment**: Always clean up system resources (SSM sessions, local ports) on exit.

> "Wormhole is not just a tool; it's the operational brain in the terminal."
