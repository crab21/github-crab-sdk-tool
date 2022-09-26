package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v47/github"
	yaml "gopkg.in/yaml.v3"
)

type T struct {
	URL           string `yaml:"url"`
	BeforeVersion string `yaml:"before_version"`
	Version       string `yaml:"version"`
	RepoName      string `yaml:"repo_name"`
	Owner         string `yaml:"owner"`
	ChangeLog     string `yaml:"change_log"`
}

type V struct {
	Value []*T `yaml:"value"`
}

func main() {
	var t = new(V)
	data, err := ioutil.ReadFile("version.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}
	err = yaml.Unmarshal(data, t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%+v", t)
	client := github.NewClient(nil)
	changeT := make([]*T, 0)
	for _, v := range t.Value {
		fmt.Println(v.Owner, v.RepoName)
		// list all organizations for user "willnorris"
		orgs, _, _ := client.Repositories.ListReleases(
			context.Background(), v.Owner, v.RepoName, &github.ListOptions{
				Page:    1,
				PerPage: 1,
			})

		if len(orgs) == 0 {
			continue
		}
		if *(orgs[0].TagName) != v.Version {
			v.BeforeVersion = v.Version
			v.Version = *(orgs[0].TagName)

			changeT = append(changeT, &T{
				RepoName:      v.RepoName,
				Version:       v.Version,
				Owner:         v.Owner,
				BeforeVersion: v.BeforeVersion,
				ChangeLog:     *(orgs[0]).Body,
			})
		}
	}
	out, _ := yaml.Marshal(t)
	ioutil.WriteFile("version.yml", out, 0666)
	changeStrs := make([]string, 0)
	changeStrs = append(changeStrs, "\n")
	changeStrs = append(changeStrs, "# All link:")
	for _, v := range t.Value {
		changeStrs = append(changeStrs, "URL: https://github.com/"+v.Owner+"/"+v.RepoName+"  \tVersion:"+v.Version)
	}

	if len(changeT) > 0 {
		changeStrs = append(changeStrs, "\n# Change link:")
		for _, v := range changeT {
			changeStrs = append(changeStrs, "\n\n")
			s := "URL: https://github.com/" + v.Owner + "/" + v.RepoName + "  \tBeforeVersion:"
			if v.BeforeVersion == "" {
				s += "None "
			} else {
				s += v.BeforeVersion
			}

			s += "  \tNowVersion:" + v.Version
			changeStrs = append(changeStrs, s)

			changeStrs = append(changeStrs, "\n"+v.RepoName+" ChangeLog:")
			changeStrs = append(changeStrs, v.ChangeLog)
		}
		ioutil.WriteFile("changet", []byte("true"), 0666)
	}

	resulstr := strings.Join(changeStrs, "\n")
	f, _ := os.OpenFile("change", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	f.WriteString(resulstr)
}
