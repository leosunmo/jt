# jt

Tiny command-line tool for creating JIRA tickets with just a Summary/Title, no description (yet).

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

Then run the command:
```bash
jt My new ticket
```

### Setting up JIRA API access
The first time you run it, it will prompt for an access token for JIRA.
You can generate one at https://id.atlassian.com/manage-profile/security/api-tokens. 

It will be stored in your system's keyring, so you won't have to enter it again until you restart or lock your keychain.