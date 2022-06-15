package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type Maintainers []*Maintainer
type Maintainer struct {
	Name    string  `yaml:"name"`
	Contact Contact `yaml:"contact"`
	Charts  []Chart `yaml:"charts"`
}

func (m Maintainers) Print() {
	for _, i := range m {
		fmt.Println(i.String())
	}
}

func (m Maintainer) String() string {
	return fmt.Sprintf("{Name: %s, Contact: %v, Charts: %v}\n", m.Name, m.Contact, m.Charts)
}

type Contact struct {
	Email        string `yaml:"email"`
	SlackChannel string `yaml:"slackChannel,omitempty"`
	URL          string `yaml:"url,omitempty"`
}

type Chart struct {
	Name          string   `yaml:"name"`
	GenerateIssue bool     `yaml:"generateIssue"`
	GithubLabels  []string `yaml:"githubLabels"`
}

type IndexFile struct {
	Entries map[string]interface{} `yaml:"entries"`
}

func main() {
	maintainersFilePath := "/Users/steven/Desktop/maintainers.yaml"
	indexFilePath := "./charts/index.yaml"
	if err := validateMaintainersFile(maintainersFilePath, indexFilePath); err != nil {
		fmt.Println(err)
	}
}

func validateMaintainersFile(maintainersFilePath, indexFilePath string) error {
	maintainers, err := decodeMaintainersFile(maintainersFilePath)
	if err != nil {
		fmt.Println(err)
	}
	// maintainers.Print()

	// Build map of charts from maintainers file and validate it there are no chart or label duplicates
	maintainersCharts := make(map[string]struct{})
	duplicateCharts := make(map[string]struct{})
	for _, m := range maintainers {
		for _, chart := range m.Charts {
			// Validate crd charts do not have generateIssue == true since we don't track crd charts on issues separately
			if strings.HasSuffix(chart.Name, "-crd") && chart.GenerateIssue {
				fmt.Printf("error: crd chart [%s] has field [generateIssue: %t] which is incorrect as crd charts are not tracked in issues separately \n", chart.Name, chart.GenerateIssue)
			}
			// Validate each chart does not have any GitHub label duplicates
			duplicateLabels := make(map[string]struct{})
			for _, label := range chart.GithubLabels {
				if _, ok := duplicateLabels[label]; ok {
					fmt.Printf("error: chart [%s] has duplicate label [%s]\n", chart.Name, label)
				}
				duplicateLabels[label] = struct{}{}
			}
			// Validate maintainers do not have any chart duplicates in their team or accross teams
			if _, ok := maintainersCharts[chart.Name]; ok {
				if _, ok := duplicateCharts[chart.Name]; !ok {
					fmt.Printf("error: chart [%s] is a duplicate or wrongly set as maintained by more than one team\n", chart.Name)
					duplicateCharts[chart.Name] = struct{}{}
				}
			}
			maintainersCharts[chart.Name] = struct{}{}
		}
	}
	index, err := decodeIndexFile(indexFilePath)
	if err != nil {
		fmt.Println(err)
	}
	if len(index.Entries) == 0 {
		fmt.Println("error: index file [%s] has no chart entries", indexFilePath)
	}

	// Validate all charts in the index file exist in the maintainers file
	for chartName := range index.Entries {
		if _, ok := maintainersCharts[chartName]; !ok {
			fmt.Printf("error: chart [%s] is missing from maintainers file [%s]\n", chartName, maintainersFilePath)
		}
	}

	// Validate all charts in the maintainers file exist in the index file
	for chartName := range maintainersCharts {
		if _, ok := index.Entries[chartName]; !ok {
			fmt.Printf("error: chart [%s] does not exist in index file [%s]\n", chartName, indexFilePath)
		}
		delete(index.Entries, chartName)
	}

	// index, err := decodeIndexFile(indexFilePath)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for chartName := range index.Entries {
	// 	if _, ok := maintainersCharts[chartName]; !ok {
	// 		fmt.Printf("error: chart %q is missing from maintainers file %s\n", chartName, maintainersFilePath)
	// 	}
	// }

	// var in IndexFile
	// file, _ := os.Open(indexFilePath)
	// defer file.Close()
	// _ = decodeYAMLFile(file, &in)
	// fmt.Printf("%v\n", in)

	// assetsPath := "./charts/assets"
	// assetsDirs, err := os.ReadDir(assetsPath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, d := range assetsDirs {
	// 	assetName := d.Name()
	// 	if strings.EqualFold(assetName, "logos") {
	// 		continue
	// 	}
	// 	if _, ok := maintainersCharts[assetName]; !ok {
	// 		fmt.Printf("error: chart %q is missing from maintainers file %s\n", assetName, path)
	// 	}
	// }

	return nil
}

func decodeMaintainersFile(path string) (Maintainers, error) {
	var maintainers Maintainers
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := decodeYAMLFile(file, &maintainers); err != nil {
		return nil, err
	}
	return maintainers, nil
}

func decodeIndexFile(path string) (*IndexFile, error) {
	var index IndexFile
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := decodeYAMLFile(file, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

// func decodeIndexFile(path string) (*repo.IndexFile, error) {
// 	var index repo.IndexFile
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()
// 	if err := decodeYAMLFile(file, &index); err != nil {
// 		return nil, err
// 	}
// 	return &index, nil
// }

func decodeYAMLFile(r io.Reader, target interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}
