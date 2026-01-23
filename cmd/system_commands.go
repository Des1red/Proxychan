package cmd

import (
	"bufio"
	"fmt"
	"os"
	"proxychan/internal/system"
	"strings"

	"golang.org/x/term"
)

func runAddUser() {
	username := prompt("Username")
	pass1 := promptPassword("Password")
	pass2 := promptPassword("Confirm password")

	if pass1 != pass2 {
		fmt.Println("passwords do not match")
		os.Exit(1)
	}

	if err := system.AddUser(username, pass1); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("user added:", username)
}

func runListUsers() {
	users, err := system.ListUsers()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if len(users) == 0 {
		fmt.Println("no users defined")
		return
	}

	for _, u := range users {
		fmt.Println("-", u)
	}
}

func runDeleteUser(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: proxychan delete-user <username>")
		os.Exit(1)
	}

	if err := system.DeleteUser(args[0]); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	} else {
		fmt.Printf("User %s deleted.", args[0])
	}
}

func prompt(label string) string {
	fmt.Print(label + ": ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptPassword(label string) string {
	fmt.Print(label + ": ")
	bytePwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("failed to read password")
		os.Exit(1)
	}
	return string(bytePwd)
}
