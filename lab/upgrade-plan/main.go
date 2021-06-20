package main

import (
	"fmt"
	"os"
	"strconv"
)

type DisruptionBudget struct {
	AppName           string
	DisruptionAllowed int
}

type Node struct {
	NodeName string
}

// Application represents an instance, like a Pod
type Application struct {
	AppName  string
	NodeName string
}

// Upgrade generates an upgrade plan, example:
//
// nodes: n1, n2, n3
// pods: [app1, n1], [app1, n2], [app2, n1], [app2, n2], [app3, n2], [app3, n3]
// budgets: [app1, 1], [app2, 1], [app3, 1]
//
// plan: [[n1, n3], [n2]]
func Upgrade(nodes []Node, pods []Application, budgets []DisruptionBudget) [][]string {
	var plan [][]string
	nodesLeft := nodes
	for len(nodesLeft) > 0 {
		// find most nodes that can be upgraded at once
		var step []string
		for _, node := range nodes {
			step = append(step, node.NodeName) // FIXME
		}
		plan = append(plan, step)
		nodesLeft = nil
	}
	return plan
}

type Dataset struct {
	Nodes   []Node
	Pods    []Application
	Budgets []DisruptionBudget
}

var datasets = []Dataset{
	{
		Nodes: []Node{
			{NodeName: "n1"},
		},
		Pods: []Application{
			{AppName: "app1", NodeName: "n1"},
		},
		Budgets: []DisruptionBudget{
			{AppName: "app1", DisruptionAllowed: 1},
		},
	},
}

// main
//
// To be considered:
// - pods may be changed (scaled, rescheduled) during operation;
// - apps may have more complex disruption restrictions;
func main() {
	var ns string
	if len(os.Args) > 1 {
		ns = os.Args[1]
	}
	if ns == "" {
		for i, dataset := range datasets {
			plan := Upgrade(dataset.Nodes, dataset.Pods, dataset.Budgets)
			fmt.Printf(
				"dataset: %d\n  nodes:\n    %v\n  pods:\n    %v\n  budgets:\n    %v\n  plan:\n",
				i, dataset.Nodes, dataset.Pods, dataset.Budgets)
			for j, step := range plan {
				fmt.Printf("    step %d: %v\n", j, step)
			}
		}
	} else {
		n, err := strconv.Atoi(ns)
		if err != nil {
			fmt.Printf("undefined dataset: %s\n", ns)
			return
		}

		if n < 0 || n > len(datasets)-1 {
			fmt.Printf("undefined dataset: %s\n", ns)
			return
		}
		dataset := datasets[n]
		plan := Upgrade(dataset.Nodes, dataset.Pods, dataset.Budgets)
		fmt.Printf("nodes:\n  %v\npods:\n  %v\nbudgets:\n  %v\nplan:\n",
			dataset.Nodes, dataset.Pods, dataset.Budgets)
		for j, step := range plan {
			fmt.Printf("  step %d: %v\n", j, step)
		}
	}
}
