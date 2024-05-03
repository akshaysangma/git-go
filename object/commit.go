package object

import (
	"fmt"
)

type Commit struct {
	Tree          string
	Parent        string
	Message       string
	Commiter      string
	CommiterEmail string
	Timestamp     string
}

func (c *Commit) Commit() (string, error) {
	content := []byte(c.Content())
	data := []byte(fmt.Sprintf("commit %d%s%s", len(content), string(nullByte), content))
	objectID, err := generateHash(data)
	if err != nil {
		return "", err
	}

	err = createObjectFile(objectID, data)
	if err != nil {
		return "", err
	}

	return objectID, nil
}

func (c *Commit) Content() string {
	return fmt.Sprintf("tree %s\nparent %s\nauthor %s <%s> %s\ncommitter %[3]s <%[4]s> %[5]s\n\n%s\n", c.Tree, c.Parent, c.Commiter, c.CommiterEmail, c.Timestamp, c.Message)
}
