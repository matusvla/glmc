# GitLab multi-cloner

This tool can clone for you all the repos in a GitLab group.


Available flags:

| Flag | Type | Usage |
| --- | --- | --- |
  -a | string | GitLab instance address (default "https://gitlab.com")
  -d | string | Root directory where the repos should be cloned
  -dry | bool | Dry run
  -g | string | GitLab group name
  -q	| bool  | Quieter output
  -t | string | Authentication token to connect to GitLab
  -v | bool.  | Verbose output of git command running under the hood
