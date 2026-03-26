package cli

import (
	"flag"
	"fmt"
	"os"
)

const rootDescription = "pidrive — Private file storage for AI agents. Files on S3, agents use ls, grep, cat. Share via URLs."

func Execute() {
	args := os.Args[1:]
	if len(args) == 0 {
		printRootUsage()
		return
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "register":
		runRegister(rest)
	case "login":
		runLogin(rest)
	case "verify":
		runVerify(rest)
	case "whoami":
		runWhoami(rest)
	case "mount":
		runMount(rest)
	case "unmount":
		runUnmount(rest)
	case "status":
		runStatus(rest)
	case "share":
		runShare(rest)
	case "shared":
		runShared(rest)
	case "revoke":
		runRevoke(rest)
	case "pull":
		runPull(rest)
	case "search":
		runSearch(rest)
	case "activity":
		runActivity(rest)
	case "trash":
		runTrash(rest)
	case "restore":
		runRestore(rest)
	case "usage":
		runUsage(rest)
	case "plans":
		runPlans(rest)
	case "upgrade":
		runUpgrade(rest)
	case "help", "--help", "-h":
		printRootUsage()
	default:
		fatalf("unknown command: %s", cmd)
	}
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {}
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✗ "+format+"\n", args...)
	os.Exit(1)
}

func printRootUsage() {
	fmt.Println(rootDescription)
	fmt.Println()
	fmt.Println("Usage: pidrive <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  register   Register a new agent account")
	fmt.Println("  login      Login to an existing account")
	fmt.Println("  verify     Verify your account with an email code")
	fmt.Println("  whoami     Show current agent info")
	fmt.Println("  mount      Mount your drive")
	fmt.Println("  unmount    Unmount your drive")
	fmt.Println("  status     Show mount and connection status")
	fmt.Println("  share      Share a file with another agent or create a link")
	fmt.Println("  shared     List all shares")
	fmt.Println("  revoke     Revoke a share")
	fmt.Println("  pull       Download a shared file")
	fmt.Println("  search     Search across your files and shared files")
	fmt.Println("  activity   Show recent activity")
	fmt.Println("  trash      List or manage deleted files")
	fmt.Println("  restore    Restore a file from trash")
	fmt.Println("  usage      Show storage and bandwidth usage")
	fmt.Println("  plans      Show available plans")
	fmt.Println("  upgrade    Upgrade your plan")
}
