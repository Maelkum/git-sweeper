package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {

	if len(os.Args) != 2 {
		return errors.New("repository path is required")
	}

	path := os.Args[1]

	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("could not open repo: %w", err)
	}

	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("could not retrieve branches")
	}

	names := make([]string, 0)
	branches.ForEach(func(ref *plumbing.Reference) error {
		names = append(names, ref.Name().Short())
		return nil
	})

	// Colored writers:
	// fresh - skip
	// older than two weeks - yellow
	// older than two months - orange
	// older than six months - red

	for _, name := range names {

		ref, err := repo.Reference(plumbing.NewBranchReferenceName(name), true)
		if err != nil {
			log.Printf("could not retrieve reference for %s: %s", name, err)
			continue
		}

		commits, err := repo.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
		if err != nil {
			log.Printf("could not retrieve commmits for %s: %s", name, err)
			continue
		}

		commit, err := commits.Next()
		if err != nil {
			log.Printf("could not retrieve first commit for %s: %s", name, err)
			continue
		}

		print(name, commit)
	}

	return nil
}

func print(branch string, commit *object.Commit) {
	ct := commit.Author.When
	fresh := calcFreshness(ct)
	message := fmtMessage(commit.Message)

	text := fmt.Sprintf("%-32s\t%s\t%s\n", branch, ct.Format("2006-01-02 15:04"), message)

	switch fresh {
	case fresh:
		color.Green(text)
	case twoWeeksPlus:
		color.Yellow(text)
	case twoMonthsPlus:
		color.RGB(255, 128, 0).Print(text)
	case sixMonthsPlus:
		color.Red(text)
	}
}

func fmtMessage(message string) string {
	lines := strings.Split(message, "\n")
	// If multiple lines - get the first one.
	if len(lines) > 1 {
		return lines[0]
	}

	// If too long - trim.
	line := lines[0]
	if len(line) > 64 {
		line = line[:64]
		line += "..."
	}

	return line
}

type freshness int

const (
	fresh freshness = iota
	twoWeeksPlus
	twoMonthsPlus
	sixMonthsPlus
)

func calcFreshness(t time.Time) freshness {

	now := time.Now()
	var (
		month = time.Hour * 24 * 30
		week  = time.Hour * 24 * 7
	)

	if t.Add(6 * month).Before(now) {
		return sixMonthsPlus
	}

	if t.Add(2 * month).Before(now) {
		return twoMonthsPlus
	}

	if t.Add(2 * week).Before(now) {
		return twoWeeksPlus
	}

	return fresh
}
