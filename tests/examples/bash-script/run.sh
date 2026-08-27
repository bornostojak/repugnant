# $rPg: Bash deployment guard, bash, deployment
# $~ Refuses to run without the environment target.
test -n "${TARGET:-}" || exit 2
