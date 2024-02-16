# jt

Tiny command-line tool for creating JIRA tickets with a Summary/Title and optionally a description.

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
```

Then you can create a ticket with:
```bash
# Create ticket with only a summary
jt My new ticket

# Create ticket with a summary and a description
jt My new ticket -m "With a description!"

# Or create a ticket with $EDITOR
jt
# The first lin is the ticket summary/title
#
# The description is everything after a blank line
# which can be multiline.
```

### gitcommit-style vim highlighting
Add this to your `.vimrc` to get gitcommit-style highlighting for the summary and description:
```vim
" jt syntax highlighting
au BufReadPost *.jt set syntax=gitcommit
```

### Setting up JIRA API access
The first time you run it, it will prompt for an access token for JIRA.
You can generate one at https://id.atlassian.com/manage-profile/security/api-tokens. 

It will be stored in your system's keyring, so you won't have to enter it again until you restart or lock your keychain.