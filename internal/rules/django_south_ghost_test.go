package rules

import "testing"

// djangoSouthGhostOutput is the (shortened) output of the Python test
// fixture; only the ghost migrations hint is relevant for the rule.
const djangoSouthGhostOutput = `Traceback (most recent call last):
  File "/home/nvbn/work/.../app/manage.py", line 34, in <module>
    execute_from_command_line(sys.argv)
  File "/home/nvbn/work/.../app/lib/south/migration/__init__.py", line 88, in check_migration_histories
    raise exceptions.GhostMigrations(ghosts)
south.exceptions.GhostMigrations:

 ! These migrations are in the database but not on disk:
    <app1: 0033_auto__...>
    <app1: 0034_fill_...>
    <app1: 0035_rename_...>
    <app2: 0003_add_...>
    <app2: 0004_denormalize_...>
 ! I'm not trusting myself; either fix this yourself by fiddling
 ! with the south_migrationhistory table, or pass --delete-ghost-migrations
 ! to South to have it delete ALL of these records (this may not be good).
`

func TestRuleDjangoSouthGhost(t *testing.T) {
	resetTestState(t)
	rule := Get("django_south_ghost")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "./manage.py migrate", djangoSouthGhostOutput)
		assertMatch(t, rule, "python manage.py migrate", djangoSouthGhostOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "./manage.py migrate", "")
		assertNotMatch(t, rule, "app migrate", djangoSouthGhostOutput)
		assertNotMatch(t, rule, "./manage.py test", djangoSouthGhostOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "./manage.py migrate auth", "",
			[]string{"./manage.py migrate auth --delete-ghost-migrations"})
	})
}
