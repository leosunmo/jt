# jt

Tiny command-line tool for creating JIRA issues with a summary, description and optionally a parent.

## Installation
If you have go install locally, compile and install with:
```bash
go install github.com/leosunmo/jt/cmd/jt
```
Otherwise download a release from the [releases page](https://github.com/leosunmo/jt/releases) and put it in your path.

## Usage
Create a config file under `~/.config/jt/config.yaml` with some default values. Here's an example with all supported values:
```yaml
url: https://example.atlassian.net
email: me@example.com
defaultProjectKey: PRJ
defaultIssueType: Task
defaultComponentNames:
  - Team A
  - Development
DefaultParentIssueTypes:
  - Epic
  - Initiative
```

Then you can create an issue with:
```bash
# Create issue with only a summary
jt My new issue

# Create issue with a summary and a description
jt My new issue -m "With a description!"

# Or create an issue with $EDITOR
jt
# The first line is the issue summary/title
#
# The description is everything after a blank line
# which can be multiline.
```

If you want to add the issue to a parent Epic or Initiative, use `-p`:
```bash
jt -p ABC-12345 Add a feature
```

### Setting up JIRA API access
The first time you run it, it will prompt for an access token for JIRA.
You can generate one at https://id.atlassian.com/manage-profile/security/api-tokens. 

It will be stored in your system's keyring, so you won't have to enter it again until you restart or lock your keychain.

### gitcommit-style vim highlighting
Add this to your `.vimrc` to get gitcommit-style highlighting for the summary and description:
```vim
" jt syntax highlighting
au BufReadPost *.jt set syntax=gitcommit
```

### auto-completion and in-line search for parent issues
The provided completion script for `zsh` allows you to not only get completion for available flags, but also
to search easily for parent issues with descriptions.

```bash
jt -p <TAB>
PROJ-12345 [Initiative]: Initiative A
PROJ-12344 [Initiative]: Initiative B
PROJ-12343 [Epic]: Epic A
PROJ-12342 [Epic]: Epic B

# You can also search for a specific description by typing it after `-p`
jt -p image<TAB>
```bash
PROJ-12340 [Initiative]: Immutable Docker Images 
PROJ-12341 [Initiative]: Image Storage Project
PROJ-12338 [Epic]: Scan Images
PROJ-12339 [Epic]: Optimize Go image

# Or search for a specific issue key
jt -p PROJ-123<TAB>
PROJ-12346 [Initiative]: Initiative C
PROJ-12347 [Initiative]: Initiative D
PROJ-12348 [Epic]: Epic C
PROJ-12349 [Epic]: Epic D
```

To enable `zsh` completion, run the following:
```bash
source <(jt --completion)
```
> **Note:** Currently only `zsh` is supported. If you want to add support for another shell, feel free to open a PR.

### TODO
- [ ] Add support for creating sub-Tasks with Tasks as parents. Currently we don't know what type the issue passed to `-p` is.