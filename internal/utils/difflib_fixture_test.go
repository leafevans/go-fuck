// The fixture tables below were generated once with Python 3.12 as a
// reference oracle. They are committed as plain Go data, so the tests never
// call Python. The values are the exact output of
// difflib.SequenceMatcher(...).ratio() and difflib.get_close_matches(...).

package utils

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

// difflibRatioFixtures holds Python's difflib.SequenceMatcher(None, left, right).ratio().
var difflibRatioFixtures = []struct {
	left  string
	right string
	want  float64
}{
	{"", "", 1.0},
	{"", "a", 0.0},
	{"a", "", 0.0},
	{"a", "a", 1.0},
	{"a", "b", 0.0},
	{"ab", "ba", 0.5},
	{"abc", "cba", 0.3333333333333333},
	{"st", "status", 0.5},
	{"st", "reset", 0.5714285714285714},
	{"st", "stage", 0.5714285714285714},
	{"st", "stash", 0.5714285714285714},
	{"st", "stats", 0.5714285714285714},
	{"st", "stop", 0.6666666666666666},
	{"git st", "git status", 0.75},
	{"git st", "git stash", 0.8},
	{"git st", "git stage", 0.8},
	{"git brnch", "git branch", 0.9473684210526315},
	{"gti", "git", 0.6666666666666666},
	{"gti", "go", 0.4},
	{"brew instal", "brew install", 0.9565217391304348},
	{"brew instal", "brew reinstall", 0.88},
	{"apt-get instol", "apt-get install", 0.896551724137931},
	{"apt-get updat", "apt-get update", 0.9629629629629629},
	{"cd /usr/loca/bin", "cd /usr/local/bin", 0.9696969696969697},
	{"python manag.py", "python manage.py", 0.967741935483871},
	{"python manage.py", "python3 manage.py", 0.9696969696969697},
	{"docker logs CONTAINER", "docker logs container", 0.5714285714285714},
	{"docker logs CONTAINER", "docker log CONTAINER", 0.975609756097561},
	{"kubectl get pods", "kubectl get pod", 0.967741935483871},
	{"npm install", "npm insall", 0.9523809523809523},
	{"yarn add", "yarn ad", 0.9333333333333333},
	{"pip install", "pip instal", 0.9523809523809523},
	{"pip3 instal", "pip3 install", 0.9565217391304348},
	{"sudo apt-get update", "apt-get update", 0.8484848484848485},
	{"make test", "make tests", 0.9473684210526315},
	{"make tes", "make test", 0.9411764705882353},
	{"go build", "go bulid", 0.875},
	{"go test ./...", "go test ./...", 1.0},
	{"cafe", "café", 0.75},
	{"café", "café", 1.0},
	{"café", "cafe", 0.75},
	{"café au lait", "café au lait", 1.0},
	{"\"quoted\"", "\"quoted\"", 1.0},
	{"你好", "你好吗", 0.8},
	{"你好", "再见", 0.0},
	{"안녕", "안녕하세요", 0.5714285714285714},
	{"안녕하세요", "안녕", 0.5714285714285714},
	{"λογος", "λογια", 0.6},
	{"λόγος", "λογος", 0.8},
	{"aaaa", "aaa", 0.8571428571428571},
	{"aaaaa", "aaa", 0.75},
	{"banana", "ananas", 0.8333333333333334},
	{"kitten", "sitting", 0.6153846153846154},
	{"saturday", "sunday", 0.7142857142857143},
	{"book", "back", 0.5},
	{"abc", "abc", 1.0},
	{"abcdefghij", "abcdefghij", 1.0},
	{"abcdefghij", "abcdefghi", 0.9473684210526315},
	{"the quick brown fox", "the quick brown fox jumps", 0.8636363636363636},
	{"the quick brown fox", "the quick brown fix", 0.9473684210526315},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxy", 0.975},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", 0.5},
	{"a b c d e f", "a b c d e g", 0.9090909090909091},
	{"--flag=value", "--flag:value", 0.9166666666666666},
	{"-rf", "-fr", 0.6666666666666666},
	{"./gradlew", "gradlew", 0.875},
	{"echo hello world", "echo hello word", 0.967741935483871},
	{"sed -e s/foo/bar", "sed -e s/foo/bar/", 0.9696969696969697},
	{"tar xvf foo.tar", "tar -xvf foo.tar", 0.967741935483871},
	{"git commit -m", "git commit --message", 0.7878787878787878},
	{"System.out.println", "system.out.print", 0.8823529411764706},
	{"docker ps -a", "docker ps", 0.8571428571428571},
	{"pip install requests", "pip install request", 0.9743589743589743},
	{"npm run build", "npm run bulid", 0.9230769230769231},
	{"yarn test", "yarn tests", 0.9473684210526315},
	{"cargo build", "cargo bulid", 0.9090909090909091},
	{"go vet ./...", "go vet ./..", 0.9565217391304348},
	{"make install", "make instal", 0.9565217391304348},
	{"sudo systemctl start", "systemctl start", 0.8571428571428571},
	{"echo 'hello world'", "echo 'hello word'", 0.9714285714285714},
	{"git rebase master", "git rebase main", 0.8125},
	{"git push origin master", "git push origin main", 0.8571428571428571},
	{"docker logs -f", "docker log -f", 0.9629629629629629},
	{"", "你好", 0.0},
	{"café", "안녕", 0.0},
	{"git status", "git stash", 0.8421052631578947},
	{"git status", "git stage", 0.7368421052631579},
	{"git status", "git reset", 0.631578947368421},
	{"git status", "git log", 0.47058823529411764},
	{"brew install", "brew uninstall", 0.9230769230769231},
	{"brew install", "brew reinstall", 0.9230769230769231},
	{"brew install", "brew list", 0.7619047619047619},
	{"apt-get install", "apt-get remove", 0.5517241379310345},
	{"apt-get install", "apt install", 0.8461538461538461},
	{"apt-get install", "apt-cache", 0.4166666666666667},
	{"docker logs", "docker log", 0.9523809523809523},
	{"docker logs", "docker ps", 0.8},
	{"docker logs", "docker exec", 0.6363636363636364},
	{"python manage.py", "python manage", 0.896551724137931},
	{"python manage.py", "python2 manage.py", 0.9696969696969697},
	{"café", "cafeteria", 0.46153846153846156},
	{"café", "decaf", 0.6666666666666666},
	{"안녕", "안녕히", 0.8},
	{"안녕", "잘가", 0.0},
	{"git commit", "git commit --amend", 0.7142857142857143},
	{"git commit", "git commit -m", 0.8695652173913043},
	{"git commit", "git committer", 0.8695652173913043},
	{"npm install", "npm install --save", 0.7586206896551724},
	{"npm install", "yarn install", 0.782608695652174},
	{"kubectl get", "kubectl get pods", 0.8148148148148148},
	{"kubectl get", "kubectl describe", 0.6666666666666666},
	{"kubectl get", "kubectl apply", 0.6666666666666666},
	{"yarn test", "yarn start", 0.7368421052631579},
	{"yarn test", "npm test", 0.7058823529411765},
	{"go test", "go test ./...", 0.7},
	{"go test", "go build", 0.4},
	{"go test", "gotest", 0.9230769230769231},
	{"git diff", "git diff --staged", 0.64},
	{"git diff", "git dif", 0.9333333333333333},
	{"git diff", "git status", 0.4444444444444444},
	{"docker run", "docker run -it", 0.8333333333333334},
	{"docker run", "docker rmi", 0.8},
	{"docker run", "docker exec", 0.6666666666666666},
}

// TestDifflibSequenceRatioPythonFixtures compares SequenceRatio with Python.
func TestDifflibSequenceRatioPythonFixtures(t *testing.T) {
	for index, fixture := range difflibRatioFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			got := SequenceRatio(fixture.left, fixture.right)
			if math.Abs(got-fixture.want) > 1e-9 {
				t.Errorf("SequenceRatio(%q, %q) = %v, want %v", fixture.left, fixture.right, got, fixture.want)
			}
		})
	}
}

// difflibCloseMatchesFixtures holds Python's difflib.get_close_matches output.
var difflibCloseMatchesFixtures = []struct {
	word          string
	possibilities []string
	n             int
	cutoff        float64
	want          []string
}{
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 1, 0.1, []string{"st"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 1, 0.5, []string{"st"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 1, 0.6, []string{"st"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 3, 0.1, []string{"st", "stop", "stats"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 3, 0.5, []string{"st", "stop", "stats"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 3, 0.6, []string{"st", "stop"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 6, 0.1, []string{"st", "stop", "stats", "stash", "stage", "reset"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 6, 0.5, []string{"st", "stop", "stats", "stash", "stage", "reset"}},
	{"st", []string{"status", "reset", "stage", "stash", "stats", "stop", "st"}, 6, 0.6, []string{"st", "stop"}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 1, 0.1, []string{"tag"}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 1, 0.6, []string{}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 3, 0.1, []string{"tag", "log", "diff"}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 3, 0.6, []string{}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 5, 0.1, []string{"tag", "log", "diff", "stash", "reset"}},
	{"gti", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 5, 0.6, []string{}},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 1, 0.6, []string{"git brunch"}},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 1, 0.1, []string{"git brunch"}},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 3, 0.6, []string{"git brunch", "git branch", "git branch -a"}},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 3, 0.1, []string{"git brunch", "git branch", "git branch -a"}},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 1, 0.6, []string{"git stash"}},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 1, 0.1, []string{"git stash"}},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 3, 0.6, []string{"git stash", "git stage", "git reset"}},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 3, 0.1, []string{"git stash", "git stage", "git reset"}},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 1, 0.6, []string{"brew install"}},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 1, 0.1, []string{"brew install"}},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 3, 0.6, []string{"brew install", "brew uninstall", "brew reinstall"}},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 3, 0.1, []string{"brew install", "brew uninstall", "brew reinstall"}},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 1, 0.6, []string{"apt-get install"}},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 1, 0.1, []string{"apt-get install"}},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 3, 0.6, []string{"apt-get install", "apt-get update", "apt-get remove"}},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 3, 0.1, []string{"apt-get install", "apt-get update", "apt-get remove"}},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 1, 0.6, []string{"/usr/local/bin"}},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 1, 0.1, []string{"/usr/local/bin"}},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 3, 0.6, []string{"/usr/local/bin", "/usr/local/sbin", "cd /usr/lib"}},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 3, 0.1, []string{"/usr/local/bin", "/usr/local/sbin", "cd /usr/lib"}},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 1, 0.6, []string{"python manage.py"}},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 1, 0.1, []string{"python manage.py"}},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 3, 0.6, []string{"python manage.py", "python3 manage.py", "python manage"}},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 3, 0.1, []string{"python manage.py", "python3 manage.py", "python manage"}},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 1, 0.6, []string{"docker log CONTAINER"}},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 1, 0.1, []string{"docker log CONTAINER"}},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 3, 0.6, []string{"docker log CONTAINER"}},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 3, 0.1, []string{"docker log CONTAINER", "docker logs container", "docker ps"}},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 1, 0.6, []string{"cafe"}},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 1, 0.1, []string{"cafe"}},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 3, 0.6, []string{"cafe", "café", "decaf"}},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 3, 0.1, []string{"cafe", "café", "decaf"}},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 1, 0.6, []string{"café"}},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 1, 0.1, []string{"café"}},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 3, 0.6, []string{"café", "cafe", "decaf"}},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 3, 0.1, []string{"café", "cafe", "decaf"}},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 1, 0.6, []string{"안녕"}},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 1, 0.1, []string{"안녕"}},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 3, 0.6, []string{"안녕", "안녕히"}},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 3, 0.1, []string{"안녕", "안녕히", "안녕하세요"}},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 1, 0.6, []string{"λογος"}},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 1, 0.1, []string{"λογος"}},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 3, 0.6, []string{"λογος", "λογο", "λογια"}},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 3, 0.1, []string{"λογος", "λογο", "λογια"}},
	{"你好", []string{"你好吗", "再见", "你好"}, 1, 0.6, []string{"你好"}},
	{"你好", []string{"你好吗", "再见", "你好"}, 1, 0.1, []string{"你好"}},
	{"你好", []string{"你好吗", "再见", "你好"}, 3, 0.6, []string{"你好", "你好吗"}},
	{"你好", []string{"你好吗", "再见", "你好"}, 3, 0.1, []string{"你好", "你好吗"}},
	{"sed -e s/foo/bar", []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}, 1, 0.6, []string{"sed -e s/foo/bar/"}},
	{"sed -e s/foo/bar", []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}, 1, 0.1, []string{"sed -e s/foo/bar/"}},
	{"sed -e s/foo/bar", []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}, 3, 0.6, []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}},
	{"sed -e s/foo/bar", []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}, 3, 0.1, []string{"sed -e s/foo/bar/", "sed -e s/foo/baz", "sed s/foo/bar"}},
	{"tar xvf foo.tar", []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}, 1, 0.6, []string{"tar -xvf foo.tar"}},
	{"tar xvf foo.tar", []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}, 1, 0.1, []string{"tar -xvf foo.tar"}},
	{"tar xvf foo.tar", []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}, 3, 0.6, []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}},
	{"tar xvf foo.tar", []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}, 3, 0.1, []string{"tar -xvf foo.tar", "tar xf foo.tar", "tar czf foo.tar"}},
	{"System.out.println", []string{"system.out.print", "System.out.print", "System.err.println"}, 1, 0.6, []string{"System.out.print"}},
	{"System.out.println", []string{"system.out.print", "System.out.print", "System.err.println"}, 1, 0.1, []string{"System.out.print"}},
	{"System.out.println", []string{"system.out.print", "System.out.print", "System.err.println"}, 3, 0.6, []string{"System.out.print", "system.out.print", "System.err.println"}},
	{"System.out.println", []string{"system.out.print", "System.out.print", "System.err.println"}, 3, 0.1, []string{"System.out.print", "system.out.print", "System.err.println"}},
	{"kubectl get pods", []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}, 1, 0.6, []string{"kubectl get pod"}},
	{"kubectl get pods", []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}, 1, 0.1, []string{"kubectl get pod"}},
	{"kubectl get pods", []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}, 3, 0.6, []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}},
	{"kubectl get pods", []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}, 3, 0.1, []string{"kubectl get pod", "kubectl get nodes", "kubectl describe pod"}},
	{"npm run build", []string{"npm run bulid", "npm run test", "npm build"}, 1, 0.6, []string{"npm run bulid"}},
	{"npm run build", []string{"npm run bulid", "npm run test", "npm build"}, 1, 0.1, []string{"npm run bulid"}},
	{"npm run build", []string{"npm run bulid", "npm run test", "npm build"}, 3, 0.6, []string{"npm run bulid", "npm build", "npm run test"}},
	{"npm run build", []string{"npm run bulid", "npm run test", "npm build"}, 3, 0.1, []string{"npm run bulid", "npm build", "npm run test"}},
	{"make test", []string{"make tests", "make install", "make test-all"}, 1, 0.6, []string{"make tests"}},
	{"make test", []string{"make tests", "make install", "make test-all"}, 1, 0.1, []string{"make tests"}},
	{"make test", []string{"make tests", "make install", "make test-all"}, 3, 0.6, []string{"make tests", "make test-all", "make install"}},
	{"make test", []string{"make tests", "make install", "make test-all"}, 3, 0.1, []string{"make tests", "make test-all", "make install"}},
	{"docker ps -a", []string{"docker ps", "docker ps -q", "docker images"}, 1, 0.6, []string{"docker ps -q"}},
	{"docker ps -a", []string{"docker ps", "docker ps -q", "docker images"}, 1, 0.1, []string{"docker ps -q"}},
	{"docker ps -a", []string{"docker ps", "docker ps -q", "docker images"}, 3, 0.6, []string{"docker ps -q", "docker ps", "docker images"}},
	{"docker ps -a", []string{"docker ps", "docker ps -q", "docker images"}, 3, 0.1, []string{"docker ps -q", "docker ps", "docker images"}},
	{"the quick brown fox", []string{"the quick brown fox jumps", "the quick brown fix", "a quick brown fox"}, 1, 0.6, []string{"the quick brown fix"}},
	{"the quick brown fox", []string{"the quick brown fox jumps", "the quick brown fix", "a quick brown fox"}, 1, 0.1, []string{"the quick brown fix"}},
	{"the quick brown fox", []string{"the quick brown fox jumps", "the quick brown fix", "a quick brown fox"}, 3, 0.6, []string{"the quick brown fix", "a quick brown fox", "the quick brown fox jumps"}},
	{"the quick brown fox", []string{"the quick brown fox jumps", "the quick brown fix", "a quick brown fox"}, 3, 0.1, []string{"the quick brown fix", "a quick brown fox", "the quick brown fox jumps"}},
	{"--flag=value", []string{"--flag:value", "--flag=values", "-flag=value"}, 1, 0.6, []string{"--flag=values"}},
	{"--flag=value", []string{"--flag:value", "--flag=values", "-flag=value"}, 1, 0.1, []string{"--flag=values"}},
	{"--flag=value", []string{"--flag:value", "--flag=values", "-flag=value"}, 3, 0.6, []string{"--flag=values", "-flag=value", "--flag:value"}},
	{"--flag=value", []string{"--flag:value", "--flag=values", "-flag=value"}, 3, 0.1, []string{"--flag=values", "-flag=value", "--flag:value"}},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}, 1, 0.6, []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}, 1, 0.1, []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}, 3, 0.6, []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}, 3, 0.1, []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}},
	{"ax", []string{"ab", "ac", "ad", "ae", "af"}, 1, 0.1, []string{"af"}},
	{"a", []string{"ab", "ac", "ad", "ae"}, 1, 0.5, []string{"ae"}},
	{"ab", []string{"ax", "ay", "az"}, 1, 0.1, []string{"az"}},
	{"st", []string{"st1", "st2", "st3", "st4", "st5", "st6", "st7", "st8"}, 1, 0.1, []string{"st8"}},
	{"git", []string{"git", "git", "git status"}, 1, 0.1, []string{"git"}},
	{"re", []string{"rebase", "reset", "grep", "rm", "read", "red"}, 1, 0.1, []string{"red"}},
	{"ax", []string{"ab", "ac", "ad", "ae", "af"}, 2, 0.1, []string{"af", "ae"}},
	{"a", []string{"ab", "ac", "ad", "ae"}, 2, 0.5, []string{"ae", "ad"}},
	{"ab", []string{"ax", "ay", "az"}, 2, 0.1, []string{"az", "ay"}},
	{"st", []string{"st1", "st2", "st3", "st4", "st5", "st6", "st7", "st8"}, 2, 0.1, []string{"st8", "st7"}},
	{"git", []string{"git", "git", "git status"}, 2, 0.1, []string{"git", "git"}},
	{"re", []string{"rebase", "reset", "grep", "rm", "read", "red"}, 2, 0.1, []string{"red", "read"}},
	{"ax", []string{"ab", "ac", "ad", "ae", "af"}, 3, 0.1, []string{"af", "ae", "ad"}},
	{"a", []string{"ab", "ac", "ad", "ae"}, 3, 0.5, []string{"ae", "ad", "ac"}},
	{"ab", []string{"ax", "ay", "az"}, 3, 0.1, []string{"az", "ay", "ax"}},
	{"st", []string{"st1", "st2", "st3", "st4", "st5", "st6", "st7", "st8"}, 3, 0.1, []string{"st8", "st7", "st6"}},
	{"git", []string{"git", "git", "git status"}, 3, 0.1, []string{"git", "git", "git status"}},
	{"re", []string{"rebase", "reset", "grep", "rm", "read", "red"}, 3, 0.1, []string{"red", "read", "grep"}},
	{"ax", []string{"ab", "ac", "ad", "ae", "af"}, 5, 0.1, []string{"af", "ae", "ad", "ac", "ab"}},
	{"a", []string{"ab", "ac", "ad", "ae"}, 5, 0.5, []string{"ae", "ad", "ac", "ab"}},
	{"ab", []string{"ax", "ay", "az"}, 5, 0.1, []string{"az", "ay", "ax"}},
	{"st", []string{"st1", "st2", "st3", "st4", "st5", "st6", "st7", "st8"}, 5, 0.1, []string{"st8", "st7", "st6", "st5", "st4"}},
	{"git", []string{"git", "git", "git status"}, 5, 0.1, []string{"git", "git", "git status"}},
	{"re", []string{"rebase", "reset", "grep", "rm", "read", "red"}, 5, 0.1, []string{"red", "read", "grep", "reset", "rm"}},
	{"", []string{"", "a"}, 1, 0.6, []string{""}},
	{"", []string{"", "a"}, 3, 0.6, []string{""}},
	{"a", []string{"a"}, 1, 1.0, []string{"a"}},
	{"a", []string{"a", "b"}, 5, 1.0, []string{"a"}},
	{"zzz", []string{"status", "reset"}, 5, 0.6, []string{}},
	{"git", []string{}, 3, 0.6, []string{}},
	{"git", []string{}, 3, 0.0, []string{}},
	{"git status", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 5, 0.3, []string{"status", "stash", "bisect", "tag"}},
	{"commit", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 4, 0.6, []string{"commit"}},
	{"push", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 4, 0.6, []string{"push"}},
	{"reset", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 3, 0.6, []string{"reset", "rebase"}},
	{"checkout", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 2, 0.6, []string{"checkout"}},
	{"branch", []string{"add", "bisect", "branch", "checkout", "clone", "commit", "diff", "fetch", "log", "merge", "pull", "push", "rebase", "reset", "status", "stash", "tag"}, 10, 0.0, []string{"branch", "stash", "fetch", "rebase", "bisect", "checkout", "tag", "add", "push", "reset"}},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}, 3, 0.1, []string{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"}},
	{"aaaa", []string{"aaa", "aaaaa", "aa", "aaaaaa"}, 4, 0.1, []string{"aaaaa", "aaa", "aaaaaa", "aa"}},
}

// TestDifflibCloseMatchesPythonFixtures compares GetCloseMatches with Python.
func TestDifflibCloseMatchesPythonFixtures(t *testing.T) {
	for index, fixture := range difflibCloseMatchesFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			got := GetCloseMatches(fixture.word, fixture.possibilities, fixture.n, fixture.cutoff)
			if !slices.Equal(got, fixture.want) {
				t.Errorf("GetCloseMatches(%q, %q, %d, %v) = %q, want %q",
					fixture.word, fixture.possibilities, fixture.n, fixture.cutoff, got, fixture.want)
			}
		})
	}
}

// difflibClosestFixtures holds thefuck's get_closest() output, which wraps
// difflib.get_close_matches with n=1 and an optional fallback to the first
// possibility.
var difflibClosestFixtures = []struct {
	word          string
	possibilities []string
	cutoff        float64
	fallback      bool
	want          string
}{
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 0.6, true, "git brunch"},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 0.6, false, "git brunch"},
	{"git brnch", []string{"git branch", "git branch -a", "git brunch", "git status"}, 0.1, true, "git brunch"},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 0.6, true, "git stash"},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 0.6, false, "git stash"},
	{"git st", []string{"git status", "git stash", "git stage", "git reset"}, 0.1, true, "git stash"},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 0.6, true, "brew install"},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 0.6, false, "brew install"},
	{"brew instal", []string{"brew install", "brew uninstall", "brew reinstall", "brew list", "brew info"}, 0.1, true, "brew install"},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 0.6, true, "apt-get install"},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 0.6, false, "apt-get install"},
	{"apt-get instol", []string{"apt-get install", "apt-get remove", "apt-get update", "apt-cache"}, 0.1, true, "apt-get install"},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 0.6, true, "/usr/local/bin"},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 0.6, false, "/usr/local/bin"},
	{"cd /usr/loca/bin", []string{"/usr/local/bin", "/usr/bin", "/usr/local/sbin", "cd /usr/lib"}, 0.1, true, "/usr/local/bin"},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 0.6, true, "python manage.py"},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 0.6, false, "python manage.py"},
	{"python manag.py", []string{"python manage.py", "python3 manage.py", "python manage", "manage.py"}, 0.1, true, "python manage.py"},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 0.6, true, "docker log CONTAINER"},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 0.6, false, "docker log CONTAINER"},
	{"docker logs CONTAINER", []string{"docker logs container", "docker log CONTAINER", "docker ps", "docker exec"}, 0.1, true, "docker log CONTAINER"},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 0.6, true, "cafe"},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 0.6, false, "cafe"},
	{"cafe", []string{"café", "cafe", "cafeteria", "decaf"}, 0.1, true, "cafe"},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 0.6, true, "café"},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 0.6, false, "café"},
	{"café", []string{"cafe", "cafeteria", "decaf", "café"}, 0.1, true, "café"},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 0.6, true, "안녕"},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 0.6, false, "안녕"},
	{"안녕", []string{"안녕하세요", "안녕히", "잘가", "안녕"}, 0.1, true, "안녕"},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 0.6, true, "λογος"},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 0.6, false, "λογος"},
	{"λογος", []string{"λογια", "λογο", "λογος", "λογική"}, 0.1, true, "λογος"},
	{"你好", []string{"你好吗", "再见", "你好"}, 0.6, true, "你好"},
	{"你好", []string{"你好吗", "再见", "你好"}, 0.6, false, "你好"},
	{"你好", []string{"你好吗", "再见", "你好"}, 0.1, true, "你好"},
	{"ax", []string{"ab", "ac", "ad"}, 0.1, true, "ad"},
	{"ax", []string{"ab", "ac", "ad"}, 0.1, false, "ad"},
	{"ab", []string{"ax", "ay", "az"}, 0.5, true, "az"},
	{"ab", []string{"ax", "ay", "az"}, 0.9, true, "ax"},
	{"ab", []string{"ax", "ay", "az"}, 0.9, false, ""},
	{"git brnch", []string{"git branch", "git branch -a"}, 0.6, true, "git branch"},
	{"zzz", []string{"status", "reset"}, 0.6, true, "status"},
	{"zzz", []string{"status", "reset"}, 0.6, false, ""},
	{"st", []string{"status", "reset", "stage"}, 0.1, true, "stage"},
	{"st", []string{"status", "reset", "stage"}, 0.1, false, "stage"},
	{"cafe", []string{"café", "cafeteria"}, 0.1, true, "café"},
	{"안녕", []string{"안녕하세요", "잘가"}, 0.1, true, "안녕하세요"},
	{"", []string{"a", "b"}, 0.6, true, "a"},
}

// TestDifflibGetClosestPythonFixtures compares GetClosest with thefuck's
// get_closest() reference implementation.
func TestDifflibGetClosestPythonFixtures(t *testing.T) {
	for index, fixture := range difflibClosestFixtures {
		t.Run(fmt.Sprintf("case_%03d", index), func(t *testing.T) {
			got := GetClosest(fixture.word, fixture.possibilities, fixture.cutoff, fixture.fallback)
			if got != fixture.want {
				t.Errorf("GetClosest(%q, %q, %v, %v) = %q, want %q",
					fixture.word, fixture.possibilities, fixture.cutoff, fixture.fallback, got, fixture.want)
			}
		})
	}
}
