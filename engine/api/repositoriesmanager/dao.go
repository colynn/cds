package repositoriesmanager

import (
	"database/sql"

	"github.com/go-gorp/gorp"
	yaml "gopkg.in/yaml.v2"

	"github.com/ovh/cds/engine/api/database/gorpmapping"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/log"
)

//InsertForProject link a project with a repository manager
func InsertForProject(db gorp.SqlExecutor, proj *sdk.Project, vcsServer *sdk.ProjectVCSServer) error {
	servers, err := LoadAllForProject(db, proj.Key)
	for _, server := range servers {
		if server.Name == vcsServer.Name {
			return sdk.WithStack(sdk.ErrConflict)
		}
	}
	if err != nil {
		return err
	}

	servers = append(servers, *vcsServer)

	b1, err := yaml.Marshal(servers)
	if err != nil {
		return err
	}

	log.Debug("repositoriesmanager.InsertForProject> %s %s", proj.Key, string(b1))

	var btes []byte
	if err := gorpmapping.Encrypt(b1, &btes, []interface{}{proj.Key}); err != nil {
		return err
	}

	if _, err := db.Exec("update project set vcs_servers = $2 where projectkey = $1", proj.Key, string(btes)); err != nil {
		return sdk.WithStack(err)
	}

	proj.VCSServers = servers

	return nil
}

//UpdateForProject update the link of project with a repository manager
func UpdateForProject(db gorp.SqlExecutor, proj *sdk.Project, vcsServers []sdk.ProjectVCSServer) error {
	b1, err := yaml.Marshal(vcsServers)
	if err != nil {
		return sdk.WithStack(err)
	}

	log.Debug("repositoriesmanager.UpdateForProject> %s %s", proj.Key, string(b1))

	var btes []byte
	if err := gorpmapping.Encrypt(b1, &btes, []interface{}{proj.Key}); err != nil {
		return err
	}

	if _, err := db.Exec("update project set vcs_servers = $2 where projectkey = $1", proj.Key, string(btes)); err != nil {
		return sdk.WithStack(err)
	}

	proj.VCSServers = vcsServers

	return nil
}

//DeleteForProject unlink a project with a repository manager
func DeleteForProject(db gorp.SqlExecutor, proj *sdk.Project, vcsServer *sdk.ProjectVCSServer) error {
	servers, err := LoadAllForProject(db, proj.Key)
	for i := range servers {
		if servers[i].Name == vcsServer.Name {
			servers = append(servers[:i], servers[i+1:]...)
			break
		}
	}

	b1, err := yaml.Marshal(servers)
	if err != nil {
		return err
	}

	var btes []byte
	if err := gorpmapping.Encrypt(b1, &btes, []interface{}{proj.Key}); err != nil {
		return err
	}

	if _, err := db.Exec("update project set vcs_servers = $2 where projectkey = $1", proj.Key, string(btes)); err != nil {
		return sdk.WithStack(err)
	}

	proj.VCSServers = servers
	return nil
}

//LoadAllForProject loads all repomanager link for a project
func LoadAllForProject(db gorp.SqlExecutor, projectKey string) ([]sdk.ProjectVCSServer, error) {
	vcsServerStr := []byte{}
	if err := db.QueryRow("select vcs_servers from project where projectkey = $1", projectKey).Scan(&vcsServerStr); err != nil {
		return nil, err
	}

	if len(vcsServerStr) == 0 {
		return []sdk.ProjectVCSServer{}, nil
	}

	var clearVCSServer []byte
	if err := gorpmapping.Decrypt(vcsServerStr, &clearVCSServer, []interface{}{projectKey}); err != nil {
		return nil, err
	}
	vcsServer := []sdk.ProjectVCSServer{}

	if err := yaml.Unmarshal(clearVCSServer, &vcsServer); err != nil {
		return nil, err
	}

	return vcsServer, nil
}

//LoadForProject loads a repomanager link for a project
func LoadForProject(db gorp.SqlExecutor, projectKey, rmName string) (*sdk.ProjectVCSServer, error) {
	vcsServerStr := []byte{}
	if err := db.QueryRow("select vcs_servers from project where projectkey = $1", projectKey).Scan(&vcsServerStr); err != nil {
		return nil, err
	}

	if len(vcsServerStr) == 0 {
		return nil, sdk.WithStack(sdk.ErrNotFound)
	}

	var clearVCSServer []byte
	if err := gorpmapping.Decrypt(vcsServerStr, &clearVCSServer, []interface{}{projectKey}); err != nil {
		return nil, err
	}

	vcsServer := []sdk.ProjectVCSServer{}
	if err := yaml.Unmarshal(clearVCSServer, &vcsServer); err != nil {
		return nil, err
	}

	for _, v := range vcsServer {
		if v.Name == rmName {
			return &v, nil
		}
	}

	return nil, sdk.WithStack(sdk.ErrNotFound)
}

//InsertForApplication associates a repositories manager with an application
func InsertForApplication(db gorp.SqlExecutor, app *sdk.Application, projectKey string) error {
	query := `UPDATE application SET vcs_server = $1, repo_fullname = $2 WHERE id = $3`
	if _, err := db.Exec(query, app.VCSServer, app.RepositoryFullname, app.ID); err != nil {
		return err
	}
	return nil
}

//DeleteForApplication removes association between  a repositories manager and an application
//it deletes the corresponding line in repositories_manager_project
func DeleteForApplication(db gorp.SqlExecutor, app *sdk.Application) error {
	query := `UPDATE application SET vcs_server = '', repo_fullname = '' WHERE id = $1`
	if _, err := db.Exec(query, app.ID); err != nil {
		return err
	}
	return nil
}

//LoadLinkedApplicationNames loads applications which are linked with this repository manager name
func LoadLinkedApplicationNames(db gorp.SqlExecutor, projectKey, rmName string) (sdk.IDNames, error) {
	query := `SELECT application.id, application.name, application.description, '' AS icon
	FROM application
		JOIN project ON project.id = application.project_id
	WHERE project.projectkey = $1 AND application.vcs_server = $2`
	var idNames sdk.IDNames
	if _, err := db.Select(&idNames, query, projectKey, rmName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, sdk.WithStack(err)
	}

	return idNames, nil
}
