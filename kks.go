package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kkga/kks/kak"
)

//go:embed init.kak
var initStr string

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	// log.SetFlags(0)

	newCmd := flag.NewFlagSet("new", flag.ExitOnError)

	editCmd := flag.NewFlagSet("edit", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendCmdBuf := sendCmd.String("b", "", "send to specified buffer")
	sendCmdAll := sendCmd.Bool("a", false, "send to all sessions and clients")

	attachCmd := flag.NewFlagSet("attach", flag.ExitOnError)

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getCmdBuf := getCmd.String("b", "", "get from specified buffer")

	killCmd := flag.NewFlagSet("kill", flag.ExitOnError)
	killCmdAll := killCmd.Bool("A", false, "kill all sessions")

	envCmd := flag.NewFlagSet("env", flag.ExitOnError)
	envCmdJson := envCmd.Bool("json", false, "json output")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmdJson := listCmd.Bool("json", false, "json output")

	catCmd := flag.NewFlagSet("cat", flag.ExitOnError)
	catCmdBuf := catCmd.String("b", "", "print specified buffer")

	sCmds := []*flag.FlagSet{editCmd, sendCmd, attachCmd, getCmd, killCmd}
	cCmds := []*flag.FlagSet{editCmd, sendCmd, attachCmd, getCmd}

	var sessArg string
	var clArg string

	for _, cmd := range sCmds {
		cmd.StringVar(&sessArg, "s", "", "Kakoune session")
	}

	for _, cmd := range cCmds {
		cmd.StringVar(&clArg, "c", "", "Kakoune client")
	}

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmdArgs := os.Args[2:]

	switch os.Args[1] {
	case "new", "n":
		newCmd.Parse(cmdArgs)
	case "edit", "e":
		editCmd.Parse(cmdArgs)
	case "send", "s":
		sendCmd.Parse(cmdArgs)
	case "attach", "a":
		attachCmd.Parse(cmdArgs)
	case "get":
		getCmd.Parse(cmdArgs)
	case "kill":
		killCmd.Parse(cmdArgs)
	case "list", "l", "ls":
		listCmd.Parse(cmdArgs)
	case "cat":
		catCmd.Parse(cmdArgs)
	case "env":
		envCmd.Parse(cmdArgs)
	case "init":
		fmt.Print(initStr)
	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(1)
	}

	// create new session
	if newCmd.Parsed() {
		name := newCmd.Arg(0)

		_, err := kak.Create(name)
		check(err)
	}

	// edit file
	if editCmd.Parsed() {
		args := editCmd.Args()
		fmt.Println(args)

		fp, err := kak.NewFilepath(args)
		check(err)

		kc, err := kak.NewContext(sessArg, clArg)
		check(err)

		if err := kc.Exists(); err != nil {
			// TODO: don't create a session, just run `kak file ...`
			newSess, err := kak.Create("")
			check(err)
			kc.Session = newSess
			kak.Connect(*fp, *kc)
		} else {
			kCmd := fmt.Sprintf("edit -existing %s", fp.Name)

			if fp.Line != 0 {
				kCmd = fmt.Sprintf("%s %d", kCmd, fp.Line)
			}
			if fp.Column != 0 {
				kCmd = fmt.Sprintf("%s %d", kCmd, fp.Column)
			}

			kak.Send(kCmd, "", kc.Session, kc.Client)
		}
	}

	// attach to session
	if attachCmd.Parsed() {
		kc, err := kak.NewContext(sessArg, clArg)
		check(err)

		contextErr := kc.Exists()
		if contextErr != nil {
			log.Fatal(contextErr)
		}
		err = kak.Connect(kak.Filepath{}, *kc)
		check(err)
	}

	// send to buffer/all
	if sendCmd.Parsed() {
		args := sendCmd.Args()
		kakCommand := strings.Join(args, " ")

		switch *sendCmdAll {
		case true:
			sessions, err := kak.List()
			check(err)

			for _, s := range sessions {
				for _, cl := range s.Clients {
					err := kak.Send(kakCommand, "", s.Name, cl)
					check(err)
				}

			}
		case false:
			kc, err := kak.NewContext(sessArg, clArg)
			check(err)
			err = kak.Send(kakCommand, *sendCmdBuf, kc.Session, kc.Client)
			check(err)
		}

	}

	// get val/opt/reg/sh from session
	if getCmd.Parsed() {
		arg := getCmd.Arg(0)

		kc, err := kak.NewContext(sessArg, clArg)
		check(err)

		if kcErr := kc.Exists(); kcErr != nil {
			log.Fatal(kcErr)
		}

		out, err := kak.Get(arg, *getCmdBuf, *kc)
		check(err)

		// TODO: this path resolution needs to happen in Edit

		// if strings.Contains(arg, "buflist") {
		// 	cwd, err := os.Getwd()
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Println("CWD:", cwd)

		// kakwd, err := kak.Get("%sh{pwd}", context.session, context.client)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Println("KAKWD:", kakwd[0])

		// 	relPath, _ := filepath.Rel(cwd, kakwd[0])
		// 	if strings.HasPrefix(relPath, "home/") {
		// 		relPath = strings.Replace(relPath, "home/", "~/", 1)
		// 	}
		// 	fmt.Println("rel path:", relPath)
		// 	fmt.Println()

		// 	for i, buf := range out {
		// 		// if !strings.HasPrefix(buf, "~") && !strings.HasPrefix(buf, "*") {
		// 		// }
		// 		if !filepath.IsAbs(buf) && !strings.HasPrefix(buf, "*") {
		// 			out[i] = filepath.Join(relPath, buf)
		// 		} else {
		// 			out[i] = buf
		// 		}
		// 	}
		// }

		fmt.Println(strings.Join(out, "\n"))
	}

	// kill session/all
	if killCmd.Parsed() {
		kakCommand := "kill"

		switch *killCmdAll {
		case true:
			sessions, err := kak.List()
			check(err)

			for _, session := range sessions {
				err = kak.Send(kakCommand, "", session.Name, "")
				check(err)
			}
		case false:
			context, err := kak.NewContext(sessArg, clArg)
			check(err)

			err = kak.Send(kakCommand, "", context.Session, context.Client)
			check(err)
		}
	}

	// list sessions
	if listCmd.Parsed() {
		sessions, err := kak.List()
		check(err)

		switch *listCmdJson {
		case true:
			j, err := json.Marshal(sessions)
			check(err)
			fmt.Println(string(j))
		case false:
			for _, session := range sessions {
				if len(session.Clients) == 0 {
					fmt.Printf("%s\t%s\t%s\n", session.Name, "null", session.Dir)
				} else {
					for _, client := range session.Clients {
						if client != "" {
							fmt.Printf("%s\t%s\t%s\n", session.Name, client, session.Dir)
						} else {
							fmt.Printf("%s\t%s\t%s\n", session.Name, "null", session.Dir)
						}
					}
				}
			}
		}
	}

	// cat buffer
	if catCmd.Parsed() {
		kc, err := kak.NewContext(sessArg, clArg)
		check(err)

		buffer := *catCmdBuf
		if buffer == "" {
			if contextErr := kc.Exists(); contextErr != nil {
				log.Fatal(contextErr)
			}
			buffile, err := kak.Get("%val{buffile}", "", *kc)
			if err != nil {
				log.Fatal(err)
			}
			buffer = buffile[0]
		}

		f, err := os.CreateTemp("", "kks-tmp")
		check(err)

		defer os.Remove(f.Name())
		defer f.Close()

		ch := make(chan string)
		go kak.ReadTmp(f, ch)

		sendCmd := fmt.Sprintf("write -force %s", f.Name())
		err = kak.Send(sendCmd, buffer, kc.Session, kc.Client)
		check(err)

		output := <-ch

		fmt.Println(output)
	}

	// print context
	if envCmd.Parsed() {
		kc, err := kak.NewContext(sessArg, clArg)
		check(err)

		err = kc.Exists()
		if err != nil {
			log.Fatal(err)
		}
		kc.Print(*envCmdJson)
	}

}

func printHelp() {
	fmt.Println(`Handy Kakoune companion.

USAGE
  kks <command> [-s <session>] [-c <client>] [<args>]

COMMANDS
  new, n         create new session
  edit, e        edit file
  send, s        send command
  attach, a      attach to session
  kill, k        kill session
  ls             list sessions and clients
  get            get %{val}, %{opt} and friends
  env            print env
  init           print Kakoune definitions

ENVIRONMENT VARIABLES
  KKS_SESSION    Kakoune session
  KKS_CLIENT     Kakoune client

Use "kks <command> -h" for command usage.`)
}
