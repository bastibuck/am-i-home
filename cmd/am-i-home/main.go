package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/bastibuck/am-i-home-cli/internal/cli"
	"github.com/bastibuck/am-i-home-cli/internal/router"
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "am-i-home - check if a device is connected to your Vodafone HomeStation\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\nUsage:\n")
	fmt.Fprintf(flag.CommandLine.Output(), "  am-i-home\n")
	fmt.Fprintf(flag.CommandLine.Output(), "    Show this help\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\n  am-i-home <FLAGS> list\n")
	fmt.Fprintf(flag.CommandLine.Output(), "    Returns a list of all active devices\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\n  am-i-home <FLAGS> list-all\n")
	fmt.Fprintf(flag.CommandLine.Output(), "    Returns a list of all devices ever connected\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\n  am-i-home <FLAGS> check <MATCHER>\n")
	fmt.Fprintf(flag.CommandLine.Output(), "    Returns 'true' or 'false' and exits 0 if MATCHER is present, 1 if absent, 2 on error\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\nFlags:\n")
	flag.PrintDefaults()
}

func main() {
	routerHost := flag.String("router", "http://192.168.0.1", "router ip address")
	pass := flag.String("pass", "", "router admin password (falls back to AM_I_HOME_ROUTER_PASS env, then .env, else interactive prompt)")
	user := flag.String("user", "admin", "router admin username")

	flag.Usage = usage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "failed parsing flags:", err)
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(0)
	}

	// ensure user flag is provided
	if strings.TrimSpace(*user) == "" {
		fmt.Fprintln(os.Stderr, "--user is required")
		os.Exit(2)
	}

	// prompt for password flag if not provided
	if *pass == "" {
		// 1) Check for environment variable
		if v, ok := os.LookupEnv("AM_I_HOME_ROUTER_PASS"); ok && v != "" {
			*pass = v
		} else {
			// 2) Try to read a .env file in the current working directory
			if b, err := os.ReadFile(".env"); err == nil {
				// parse simple KEY=VALUE lines
				for line := range strings.SplitSeq(string(b), "\n") {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					parts := strings.SplitN(line, "=", 2)
					if len(parts) != 2 {
						continue
					}
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if len(val) >= 2 {
						if (val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"') {
							val = val[1 : len(val)-1]
						}
					}
					if key == "AM_I_HOME_ROUTER_PASS" && val != "" {
						*pass = val
						break
					}
				}
			}
		}

		if *pass == "" {
			// only attempt an interactive prompt when stdin is a terminal
			if term.IsTerminal(int(syscall.Stdin)) {
				fmt.Printf("Password for %s@%s: ", *user, *routerHost)
				p, err := term.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					fmt.Fprintln(os.Stderr, "failed reading password:", err)
					os.Exit(2)
				}
				*pass = string(p)
			} else {
				fmt.Fprintln(os.Stderr, "--pass is required when not running interactively and AM_I_HOME_ROUTER_PASS not set")
				os.Exit(2)
			}
		}
	}

	// create HomeStation client (uses cookiejar internally)
	hs, err := router.NewHomeStationClient(*routerHost, *user, *pass)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed creating HomeStation client:", err)
		os.Exit(2)
	}

	switch args[0] {
	case "list-all":
		if err := cli.ListDevices(hs); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}

	case "list":
		if err := cli.ListActive(hs); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}

	case "check":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "check command requires a matcher argument")
			os.Exit(2)
		}
		matcher := args[1]
		found, err := cli.CheckByMatcher(hs, matcher)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}

		fmt.Println(found)
		if found {
			os.Exit(0)
		}

		os.Exit(1)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage()
		os.Exit(2)
	}
}
