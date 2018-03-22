// Command ant is a deployment tools of ant microservice frameware.
//
// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package main

import (
	"os"

	"github.com/xiaoenai/ants/ant/create"
	"github.com/xiaoenai/ants/ant/info"
	"github.com/xiaoenai/ants/ant/run"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Ant project aids"
	app.Version = "0.1.0"
	app.Author = "henrylee2cn"
	app.Usage = "a deployment tools of ant microservice frameware"

	// new a project
	newCom := cli.Command{
		Name:  "gen",
		Usage: "Generate an ant project",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "template, t",
				Usage: "The template for code generation(relative/absolute)",
			},
			cli.StringFlag{
				Name:  "app_path, p",
				Usage: "The path(relative/absolute) of the project",
			},
		},
		Before: initProject,
		Action: func(c *cli.Context) error {
			create.CreateProject(c.String("template"))
			return nil
		},
	}

	// run a project
	runCom := cli.Command{
		Name:  "run",
		Usage: "Compile and run gracefully (monitor changes) an any existing go project",
		UsageText: `ant run [options] [arguments...]
 or
   ant run [options except -app_path] [arguments...] {app_path}`,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "watch_exts, x",
				Value: (*cli.StringSlice)(&[]string{".go", ".ini", ".yaml", ".toml", ".xml"}),
				Usage: "Specified to increase the listening file suffix",
			},
			cli.StringFlag{
				Name:  "app_path, p",
				Usage: "The path(relative/absolute) of the project",
			},
		},
		Before: initProject,
		Action: func(c *cli.Context) error {
			run.RunProject(c.StringSlice("watch_exts"))
			return nil
		},
	}

	app.Commands = []cli.Command{newCom, runCom}
	app.Run(os.Args)
}

func initProject(c *cli.Context) error {
	appPath := c.String("app_path")
	if len(appPath) == 0 {
		appPath = c.Args().First()
	}
	if len(appPath) == 0 {
		appPath = "./"
	}
	return info.Init(appPath)
}
