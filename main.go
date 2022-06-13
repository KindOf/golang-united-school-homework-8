package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func newUserFromBytes(b []byte) (*User, error) {
	var user User
	err := json.Unmarshal(b, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type Users []User

func (us Users) toBytes() ([]byte, error) {
	b, err := json.Marshal(us)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func newUsersFromBytes(input []byte) (Users, error) {
	var users Users
	err := json.Unmarshal(input, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func getUsersFromFile(file *os.File) (Users, error) {
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(byteValue) == 0 {
		byteValue = []byte("[]")
	}

	users, err := newUsersFromBytes(byteValue)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func list(fileName string) ([]byte, error) {
	if fileName == "" {
		return nil, errors.New("-fileName flag has to be specified")
	}

	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func isIdExists(users []User, id string) bool {
	for _, user := range users {
		if id == user.Id {
			return true
		}
	}
	return false
}

func add(item string, fileName string) ([]byte, error) {
	if item == "" {
		return nil, errors.New("-item flag has to be specified")
	}

	userItem, err := newUserFromBytes([]byte(item))

	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	users, err := getUsersFromFile(file)
	if err != nil {
		return nil, err
	}

	if isIdExists(users, userItem.Id) {
		return []byte(fmt.Sprintf("Item with id %v already exists", userItem.Id)), nil
	}

	users = append(users, *userItem)

	b, err := users.toBytes()
	if err != nil {
		return nil, err
	}

	_, err = file.Write(b)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func remove(id string, fileName string) ([]byte, error) {
	if id == "" {
		return nil, errors.New("-id flag has to be specified")
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	users, err := getUsersFromFile(file)
	if err != nil {
		return nil, err
	}

	index := -1

	for i, user := range users {
		if user.Id == id {
			index = i
			break
		}
	}

	if index == -1 {
		return []byte("Item with id 2 not found"), nil
	}

	us := append(users[:index], users[index+1:]...)

	b, err := us.toBytes()
	if err != nil {
		return nil, err
	}

	err = file.Truncate(0)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	_, err = file.Write(b)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func findById(fileName string, id string) ([]byte, error) {
	if id == "" {
		return nil, errors.New("-id flag has to be specified")
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	users, err := getUsersFromFile(file)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Id == id {
			bytes, err := json.Marshal(user)
			if err != nil {
				return nil, err
			}

			return bytes, nil
		}
	}
	return []byte(""), nil
}

func Perform(args Arguments, writer io.Writer) error {
	operation := args["operation"]

	var err error
	var b []byte

	switch operation {
	case "list":
		b, err = list(args["fileName"])
	case "add":
		b, err = add(args["item"], args["fileName"])
	case "remove":
		b, err = remove(args["id"], args["fileName"])
	case "findById":
		b, err = findById(args["fileName"], args["id"])
	case "":
		return errors.New("-operation flag has to be specified")
	default:
		return errors.New(fmt.Sprintf("Operation %v not allowed!", operation))
	}

	if err != nil {
		return err
	}

	_, e := writer.Write(b)
	if e != nil {
		return e
	}
	return nil
}

func parseArgs() Arguments {
	a := make(Arguments)

	id := flag.String("id", "", "User id")
	operation := flag.String("operation", "", "list, add, remove, findById")
	item := flag.String("item", "", "should be json string")
	fileName := flag.String("fileName", "users.json", "File name")

	flag.Parse()

	a["operation"] = *operation
	a["item"] = *item
	a["id"] = *id
	a["fileName"] = *fileName

	return a
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
