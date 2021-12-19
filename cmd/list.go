package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kkga/kks/kak"
)

func NewListCmd() *ListCmd {
	c := &ListCmd{Cmd: Cmd{
		fs:          flag.NewFlagSet("list", flag.ExitOnError),
		aliases:     []string{"ls", "l"},
		description: "List Kakoune sessions and clients.",
		usageLine:   "[options]",
	}}
	c.fs.BoolVar(&c.json, "json", false, "json output")
	return c
}

type ListCmd struct {
	Cmd
	json bool
}

func (c *ListCmd) Run() error {
	kakSessions, err := kak.Sessions()
	if err != nil {
		return err
	}

	if c.json {
		type session struct {
			Name    string   `json:"name"`
			Clients []string `json:"clients"`
			Dir     string   `json:"dir"`
		}

		sessions := make([]session, len(kakSessions))

		for i, s := range kakSessions {
			d, err := s.Dir()
			if err != nil {
				return err
			}

			sessions[i] = session{Name: s.Name, Clients: []string{}, Dir: d}

			clients, err := s.Clients()
			if err != nil {
				return err
			}
			for _, c := range clients {
				if c.Name != "" {
					sessions[i].Clients = append(sessions[i].Clients, c.Name)
				}
			}
		}

		j, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(j))
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)

		for _, s := range kakSessions {
			c, err := s.Clients()
			if err != nil {
				return err
			}

			d, err := s.Dir()
			if err != nil {
				return err
			}

			if len(c) == 0 {
				fmt.Fprintf(w, "%s\t: %s\t: %s\n", s.Name, " ", d)
			} else {
				for _, cl := range c {
					fmt.Fprintf(w, "%s\t: %s\t: %s\n", s.Name, cl.Name, d)
				}
			}
		}

		w.Flush()
	}

	return nil
}
