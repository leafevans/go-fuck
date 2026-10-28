package rules

import "testing"

// dockerLoginPullOutput is the output of a failing "docker build".
const dockerLoginPullOutput = `
    Sending build context to Docker daemon  118.8kB
Step 1/6 : FROM foo/bar:fdb7c6d
pull access denied for foo/bar, repository does not exist or may require 'docker login'
`

// dockerLoginPushOutput is the output of a failing "docker push".
const dockerLoginPushOutput = `
    The push refers to repository [artifactory:9090/foo/bar]
push access denied for foo/bar, repository does not exist or may require 'docker login'
`

// dockerLoginSuccessOutput is a successful "docker push" output.
const dockerLoginSuccessOutput = `
    docker push artifactory:9090/foo/bar:fdb7c6d
The push refers to repository [artifactory:9090/foo/bar]
9c29c7ad209d: Preparing
71f3ad53dfe0: Preparing
f58ee068224c: Preparing
aeddc924d0f7: Preparing
c2040e5d6363: Preparing
4d42df4f350f: Preparing
35723dab26f9: Preparing
71f3ad53dfe0: Pushed
cb95fa0faeb1: Layer already exists
`

func TestRuleDockerLogin(t *testing.T) {
	resetTestState(t)
	rule := Get("docker_login")

	t.Run("match pull access denied", func(t *testing.T) {
		assertMatch(t, rule, "docker build -t artifactory:9090/foo/bar:fdb7c6d .", dockerLoginPullOutput)
	})

	t.Run("match push access denied", func(t *testing.T) {
		assertMatch(t, rule, "docker push artifactory:9090/foo/bar:fdb7c6d", dockerLoginPushOutput)
	})

	t.Run("not match successful push", func(t *testing.T) {
		assertNotMatch(t, rule, "docker push artifactory:9090/foo/bar:fdb7c6d", dockerLoginSuccessOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "docker build -t artifactory:9090/foo/bar:fdb7c6d .", "",
			[]string{"docker login && docker build -t artifactory:9090/foo/bar:fdb7c6d ."})
		assertNewCommands(t, rule, "docker push artifactory:9090/foo/bar:fdb7c6d", "",
			[]string{"docker login && docker push artifactory:9090/foo/bar:fdb7c6d"})
	})
}
