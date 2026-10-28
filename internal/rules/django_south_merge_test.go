package rules

import "testing"

// djangoSouthMergeOutput is the (shortened) output of the Python test fixture;
// only the --merge hint is relevant for the rule.
const djangoSouthMergeOutput = `Running migrations for app:
 ! Migration app:0003_auto... should not have been applied before app:0002_auto__add_field_query_due_date_ but was.
Traceback (most recent call last):
  File "/home/nvbn/work/.../app/manage.py", line 34, in <module>
    execute_from_command_line(sys.argv)
  File "/home/nvbn/work/.../app/lib/south/migration/__init__.py", line 207, in migrate_app
    raise exceptions.InconsistentMigrationHistory(problems)
south.exceptions.InconsistentMigrationHistory: Inconsistent migration history
The following options are available:
    --merge: will just attempt the migration ignoring any potential dependency conflicts.
`

func TestRuleDjangoSouthMerge(t *testing.T) {
	resetTestState(t)
	rule := Get("django_south_merge")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "./manage.py migrate", djangoSouthMergeOutput)
		assertMatch(t, rule, "python manage.py migrate", djangoSouthMergeOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "./manage.py migrate", "")
		assertNotMatch(t, rule, "app migrate", djangoSouthMergeOutput)
		assertNotMatch(t, rule, "./manage.py test", djangoSouthMergeOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "./manage.py migrate auth", "",
			[]string{"./manage.py migrate auth --merge"})
	})
}
