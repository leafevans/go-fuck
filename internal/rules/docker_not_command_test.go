package rules

import "testing"

// dockerNotCommandOldHelp is the help output of older docker versions. The
// option list is omitted because it is not parsed. The commands section is
// copied from the Python test fixture.
const dockerNotCommandOldHelp = `Usage: docker [OPTIONS] COMMAND [arg...]

A self-sufficient runtime for linux containers.

Commands:
    attach    Attach to a running container
    build     Build an image from a Dockerfile
    commit    Create a new image from a container's changes
    cp        Copy files/folders from a container's filesystem to the host path
    create    Create a new container
    diff      Inspect changes on a container's filesystem
    events    Get real time events from the server
    exec      Run a command in a running container
    export    Stream the contents of a container as a tar archive
    history   Show the history of an image
    images    List images
    import    Create a new filesystem image from the contents of a tarball
    info      Display system-wide information
    inspect   Return low-level information on a container or image
    kill      Kill a running container
    load      Load an image from a tar archive
    login     Register or log in to a Docker registry server
    logout    Log out from a Docker registry server
    logs      Fetch the logs of a container
    pause     Pause all processes within a container
    port      Lookup the public-facing port that is NAT-ed to PRIVATE_PORT
    ps        List containers
    pull      Pull an image or a repository from a Docker registry server
    push      Push an image or a repository to a Docker registry server
    rename    Rename an existing container
    restart   Restart a running container
    rm        Remove one or more containers
    rmi       Remove one or more images
    run       Run a command in a new container
    save      Save an image to a tar archive
    search    Search for an image on the Docker Hub
    start     Start a stopped container
    stats     Display a stream of a containers' resource usage statistics
    stop      Stop a running container
    tag       Tag an image into a repository
    top       Lookup the running processes of a container
    unpause   Unpause a paused container
    version   Show the Docker version information
    wait      Block until a container stops, then print its exit code

Run 'docker COMMAND --help' for more information on a command.
`

// dockerNotCommandNewHelp is the help output of newer docker versions. It is
// written to stderr and contains management commands.
const dockerNotCommandNewHelp = `
Usage:	docker [OPTIONS] COMMAND

A self-sufficient runtime for containers

Management Commands:
  builder     Manage builds
  config      Manage Docker configs
  container   Manage containers
  context     Manage contexts
  image       Manage images
  network     Manage networks
  node        Manage Swarm nodes
  plugin      Manage plugins
  secret      Manage Docker secrets
  service     Manage services
  stack       Manage Docker stacks
  swarm       Manage Swarm
  system      Manage Docker
  trust       Manage trust on Docker images
  volume      Manage volumes

Commands:
  attach      Attach local standard input, output, and error streams to a running container
  build       Build an image from a Dockerfile
  commit      Create a new image from a container's changes
  cp          Copy files/folders between a container and the local filesystem
  create      Create a new container
  diff        Inspect changes to files or directories on a container's filesystem
  events      Get real time events from the server
  exec        Run a command in a running container
  export      Export a container's filesystem as a tar archive
  history     Show the history of an image
  images      List images
  import      Import the contents from a tarball to create a filesystem image
  info        Display system-wide information
  inspect     Return low-level information on Docker objects
  kill        Kill one or more running containers
  load        Load an image from a tar archive or STDIN
  login       Log in to a Docker registry
  logout      Log out from a Docker registry
  logs        Fetch the logs of a container
  pause       Pause all processes within one or more containers
  port        List port mappings or a specific mapping for the container
  ps          List containers
  pull        Pull an image or a repository from a registry
  push        Push an image or a repository to a registry
  rename      Rename a container
  restart     Restart one or more containers
  rm          Remove one or more containers
  rmi         Remove one or more images
  run         Run a command in a new container
  save        Save one or more images to a tar archive (streamed to STDOUT by default)
  search      Search the Docker Hub for images
  start       Start one or more stopped containers
  stats       Display a live stream of container(s) resource usage statistics
  stop        Stop one or more running containers
  tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE
  top         Display the running processes of a container
  unpause     Unpause all processes within one or more containers
  update      Update configuration of one or more containers
  version     Show the Docker version information
  wait        Block until one or more containers stop, then print their exit codes

Run 'docker COMMAND --help' for more information on a command.
`

// dockerNotCommandSwarmOutput is the usage output of "docker swarm".
const dockerNotCommandSwarmOutput = `
Usage:	docker swarm COMMAND

Manage Swarm

Commands:
  ca          Display and rotate the root CA
  init        Initialize a swarm
  join        Join a swarm as a node and/or manager
  join-token  Manage join tokens
  leave       Leave the swarm
  unlock      Unlock swarm
  unlock-key  Manage the unlock key
  update      Update the swarm

Run 'docker swarm COMMAND --help' for more information on a command.
`

// dockerNotCommandImageOutput is the usage output of "docker image".
const dockerNotCommandImageOutput = `
Usage:	docker image COMMAND

Manage images

Commands:
  build       Build an image from a Dockerfile
  history     Show the history of an image
  import      Import the contents from a tarball to create a filesystem image
  inspect     Display detailed information on one or more images
  load        Load an image from a tar archive or STDIN
  ls          List images
  prune       Remove unused images
  pull        Pull an image or a repository from a registry
  push        Push an image or a repository to a registry
  rm          Remove one or more images
  save        Save one or more images to a tar archive (streamed to STDOUT by default)
  tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE

Run 'docker image COMMAND --help' for more information on a command.
`

// dockerNotCommandOutput builds the "not a docker command" error message.
func dockerNotCommandOutput(name string) string {
	return "docker: '" + name + "' is not a docker command.\nSee 'docker --help'."
}

// dockerNotCommandUseHelp replaces the docker help lookup with canned output.
func dockerNotCommandUseHelp(t *testing.T, stdout, stderr string) {
	t.Helper()
	original := dockerNotCommandCommands
	dockerNotCommandCommands = func() []string {
		return dockerNotCommandHelpCommands(stdout, stderr)
	}
	t.Cleanup(func() { dockerNotCommandCommands = original })
}

func TestRuleDockerNotCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("docker_not_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "docker pes", dockerNotCommandOutput("pes"))
	})

	t.Run("match management command", func(t *testing.T) {
		assertMatch(t, rule, "docker swarn", dockerNotCommandOutput("swarn"))
		assertMatch(t, rule, "docker imge", dockerNotCommandOutput("imge"))
	})

	t.Run("match management subcommand", func(t *testing.T) {
		assertMatch(t, rule, "docker swarm int", dockerNotCommandSwarmOutput)
		assertMatch(t, rule, "docker image la", dockerNotCommandImageOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "docker ps", "")
		assertNotMatch(t, rule, "cat pes", dockerNotCommandOutput("pes"))
	})

	t.Run("get new command", func(t *testing.T) {
		dockerNotCommandUseHelp(t, dockerNotCommandOldHelp, "")
		assertNewCommands(t, rule, "docker pes", dockerNotCommandOutput("pes"),
			[]string{"docker ps", "docker push", "docker pause"})
		assertNewCommands(t, rule, "docker tags", dockerNotCommandOutput("tags"),
			[]string{"docker tag", "docker stats", "docker images"})
	})

	t.Run("get new management command", func(t *testing.T) {
		dockerNotCommandUseHelp(t, "", dockerNotCommandNewHelp)
		assertNewCommands(t, rule, "docker swarn", dockerNotCommandOutput("swarn"),
			[]string{"docker swarm", "docker start", "docker search"})
		assertNewCommands(t, rule, "docker inage", dockerNotCommandOutput("inage"),
			[]string{"docker image", "docker images", "docker rename"})
	})

	t.Run("get new management subcommand", func(t *testing.T) {
		dockerNotCommandUseHelp(t, dockerNotCommandOldHelp, "")
		assertNewCommands(t, rule, "docker swarm int", dockerNotCommandSwarmOutput,
			[]string{"docker swarm init", "docker swarm join", "docker swarm join-token"})
		assertNewCommands(t, rule, "docker image la", dockerNotCommandImageOutput,
			[]string{"docker image load", "docker image ls", "docker image tag"})
	})
}
