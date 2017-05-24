// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	// MPC : count of my projects
	MPC = "my_project_count"
	// MRC : count of my repositories
	MRC = "my_repo_count"
	// PPC : count of public projects
	PPC = "public_project_count"
	// PRC : count of public repositories
	PRC = "public_repo_count"
	// TPC : total count of projects
	TPC = "total_project_count"
	// TRC : total count of repositories
	TRC = "total_repo_count"
)

// StatisticAPI handles request to /api/statistics/
type StatisticAPI struct {
	BaseController
	username string
}

//Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	s.BaseController.Prepare()
	if !s.SecurityCtx.IsAuthenticated() {
		s.HandleUnauthorized()
		return
	}
	s.username = s.SecurityCtx.GetUsername()
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	statistic := map[string]int64{}

	projects, err := s.ProjectMgr.GetPublic()
	if err != nil {
		s.HandleInternalServerError(fmt.Sprintf(
			"failed to get public projects: %v", err))
		return
	}

	statistic[PPC] = (int64)(len(projects))

	ids := []int64{}
	for _, p := range projects {
		ids = append(ids, p.ProjectID)
	}
	n, err := dao.GetTotalOfRepositoriesByProject(ids, "")
	if err != nil {
		log.Errorf("failed to get total of public repositories: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}
	statistic[PRC] = n

	if s.SecurityCtx.IsSysAdmin() {
		n, err := dao.GetTotalOfProjects(nil)
		if err != nil {
			log.Errorf("failed to get total of projects: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MPC] = n
		statistic[TPC] = n

		n, err = dao.GetTotalOfRepositories("")
		if err != nil {
			log.Errorf("failed to get total of repositories: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MRC] = n
		statistic[TRC] = n
	} else {
		projects, err := s.ProjectMgr.GetAll(&models.QueryParam{
			Member: &models.Member{
				Name: s.username,
			},
		})
		if err != nil {
			s.HandleInternalServerError(fmt.Sprintf(
				"failed to get projects: %v", err))
			return
		}
		statistic[MPC] = (int64)(len(projects))

		ids := []int64{}
		for _, p := range projects {
			ids = append(ids, p.ProjectID)
		}

		n, err = dao.GetTotalOfRepositoriesByProject(ids, "")
		if err != nil {
			s.HandleInternalServerError(fmt.Sprintf(
				"failed to get total of repositories for user %s: %v",
				s.username, err))
			return
		}
		statistic[MRC] = n
	}

	s.Data["json"] = statistic
	s.ServeJSON()
}
