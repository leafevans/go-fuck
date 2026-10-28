package rules

import "testing"

// portAlreadyInUseNodeOutput is the output of a node.js server that cannot
// bind its port.
const portAlreadyInUseNodeOutput = `

DE 70% 1/1 build modulesevents.js:141
      throw er; // Unhandled 'error' event
      ^

Error: listen EADDRINUSE 127.0.0.1:8080
    at Object.exports._errnoException (util.js:873:11)
    at exports._exceptionWithHostPort (util.js:896:20)
    at Server._listen2 (net.js:1250:14)
    at listen (net.js:1286:10)
    at net.js:1395:9
    at GetAddrInfoReqWrap.asyncCallback [as callback] (dns.js:64:16)
    at GetAddrInfoReqWrap.onlookup [as oncomplete] (dns.js:83:10)

    `

// portAlreadyInUsePackagerOutput is the output of the react-native packager.
const portAlreadyInUsePackagerOutput = `
[6:40:01 AM] <START> Building Dependency Graph
[6:40:01 AM] <START> Crawling File System
 ERROR  Packager can't listen on port 8080
Most likely another process is already using this port
Run the following command to find out which process:

   lsof -n -i4TCP:8080

You can either shut down the other process:

   kill -9 <PID>

or run packager on different port.

    `

// portAlreadyInUseAiohttpOutput is the output of an aiohttp application.
const portAlreadyInUseAiohttpOutput = `
Traceback (most recent call last):
  File "/usr/lib/python3.5/runpy.py", line 184, in _run_module_as_main
    "__main__", mod_spec)
  File "/usr/lib/python3.5/runpy.py", line 85, in _run_code
    exec(code, run_globals)
  File "/home/nvbn/exp/code_view/server/code_view/main.py", line 14, in <module>
    web.run_app(app)
  File "/home/nvbn/.virtualenvs/code_view/lib/python3.5/site-packages/aiohttp/web.py", line 310, in run_app
    backlog=backlog))
  File "/usr/lib/python3.5/asyncio/base_events.py", line 373, in run_until_complete
    return future.result()
  File "/usr/lib/python3.5/asyncio/futures.py", line 274, in result
    raise self._exception
  File "/usr/lib/python3.5/asyncio/tasks.py", line 240, in _step
    result = coro.send(None)
  File "/usr/lib/python3.5/asyncio/base_events.py", line 953, in create_server
    % (sa, err.strerror.lower()))
OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8080): address already in use
Task was destroyed but it is pending!
task: <Task pending coro=<RedisProtocol._reader_coroutine() running at /home/nvbn/.virtualenvs/code_view/lib/python3.5/site-packages/asyncio_redis/protocol.py:921> wait_for=<Future pending cb=[Task._wakeup()]>>
    `

// portAlreadyInUseOutputs lists all supported error outputs.
var portAlreadyInUseOutputs = []string{
	portAlreadyInUseNodeOutput,
	portAlreadyInUsePackagerOutput,
	portAlreadyInUseAiohttpOutput,
}

// portAlreadyInUseLsofStdout is the output of lsof of the Python tests.
const portAlreadyInUseLsofStdout = `COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node    18233 nvbn   16u  IPv4 557134      0t0  TCP localhost:http-alt (LISTEN)
`

// portAlreadyInUseUsePID replaces the PID lookup with the Python fixture.
func portAlreadyInUseUsePID(t *testing.T) {
	t.Helper()
	original := portAlreadyInUsePID
	portAlreadyInUsePID = func(string) string {
		return portAlreadyInUseParsePID(portAlreadyInUseLsofStdout)
	}
	t.Cleanup(func() { portAlreadyInUsePID = original })
}

// portAlreadyInUseUseNoPID replaces the PID lookup with an empty result.
func portAlreadyInUseUseNoPID(t *testing.T) {
	t.Helper()
	original := portAlreadyInUsePID
	portAlreadyInUsePID = func(string) string { return "" }
	t.Cleanup(func() { portAlreadyInUsePID = original })
}

func TestRulePortAlreadyInUse(t *testing.T) {
	resetTestState(t)
	rule := Get("port_already_in_use")

	t.Run("match", func(t *testing.T) {
		portAlreadyInUseUsePID(t)
		for _, output := range portAlreadyInUseOutputs {
			assertMatch(t, rule, "./app", output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		portAlreadyInUseUsePID(t)
		assertNotMatch(t, rule, "./app", "")

		portAlreadyInUseUseNoPID(t)
		assertNotMatch(t, rule, "./app", portAlreadyInUseOutputs[1])
		assertNotMatch(t, rule, "./app", portAlreadyInUseOutputs[2])
	})

	t.Run("get new command", func(t *testing.T) {
		portAlreadyInUseUsePID(t)
		for _, output := range portAlreadyInUseOutputs {
			assertNewCommands(t, rule, "./app", output, []string{"kill 18233 && ./app"})
		}
	})

	t.Run("parse port", func(t *testing.T) {
		for _, output := range portAlreadyInUseOutputs {
			if port := portAlreadyInUseParsePort(output); port != "8080" {
				t.Errorf("portAlreadyInUseParsePort returned %q, want %q", port, "8080")
			}
		}
		if port := portAlreadyInUseParsePort("no port here"); port != "" {
			t.Errorf("portAlreadyInUseParsePort returned %q, want an empty string", port)
		}
	})
}
