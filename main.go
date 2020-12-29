package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func init() {
	log.SetFlags(0)
}

func walk(n *yaml.Node, depth int) {
	log.Printf("%s%+v", strings.Repeat(" ", depth), n)
	for _, child := range n.Content {
		walk(child, depth+1)
	}
}

func run() error {
	f, err := os.Open("testdata/fixture1.yaml")
	if err != nil {
		return fmt.Errorf("could not open the fixture: %w", err)
	}
	defer f.Close()

	imagePath, err := yamlpath.NewPath("$.spec.template.spec.containers[*].image")
	if err != nil {
		return fmt.Errorf("invalid yaml path: %w", err)
	}
	commandPath, err := yamlpath.NewPath("$.spec.template.spec.containers[*].command")
	if err != nil {
		return fmt.Errorf("invalid yaml path: %w", err)
	}

	d := yaml.NewDecoder(f)
	e := yaml.NewEncoder(os.Stdout)
	e.SetIndent(2)
	for {
		var n yaml.Node
		if err := d.Decode(&n); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("could not decode: %w", err)
		}

		// replace the image name
		imageNodes, err := imagePath.Find(&n)
		if err != nil {
			return fmt.Errorf("could not find image: %w", err)
		}
		for _, imageNode := range imageNodes {
			if strings.HasPrefix(imageNode.Value, "nginx:") {
				imageNode.SetString("NEW_IMAGE")
			}
		}

		// append a string to the command
		commandNodes, err := commandPath.Find(&n)
		if err != nil {
			return fmt.Errorf("could not find command: %w", err)
		}
		for _, commandNode := range commandNodes {
			commandNode.Content = append(commandNode.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "foo",
			})
		}

		// dump the tree
		walk(&n, 0)

		// write the document
		if err := e.Encode(&n); err != nil {
			return fmt.Errorf("could not encode: %w", err)
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %s", err)
	}
}
