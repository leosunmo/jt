# Zsh completion script for the 'jt' command

# Always show the list, even if there's only one option.
zstyle ':completion:*:jt:*' force-list always

# Disable sorting to preserve the custom order from jt -q.
zstyle ':completion::complete:jt::' sort false

# Define the jt completion function for Zsh
_jt_completions() {
    # Define the arguments with exclusivity
    _arguments -C \
        '(-m --msg)'{-m,--msg}'[Issue description, optional]:description' \
        '(-e --edit)'{-e,--edit}'[Open default editor for summary and description, optional]' \
        '(-p --parent)'{-p,--parent}'[Assign the issue to a parent Epic or Initiative, optional]:project:->parent_completion' \
        '(-c --completion)'{-c,--completion}'[Print zsh shell completion script to stdout and exit]' \
        '(-q --query)'{-q,--query}'[Query issues and exit. Options: parents, epics, initiatives, tasks, bugs. Comma followed by string for description search]:query:->query_completion' \
        '(-h --help)'{-h,--help}'[Show help]' &&
        return 0

    # Handle completion based on the context
    case $state in
    parent_completion)
        # Define the completion options for the -p/--parent flag

        # Enable immediate menu display and prevent sorting
        compstate[insert]=menu
        # Prevent sorting to preserve custom order
        compstate[nosort]=true

        local -a insertions descriptions
        local search_term="${words[CURRENT]}"
        local query_param="parents"
        local has_matches=0 # Flag to check if matches are found in this section

        # If the search term is empty, or an issue ID prefix (e.g., PLT-77), fetch the full list
        if [[ -z "$search_term" || "$search_term" =~ '^([[:alpha:]]{3,4})-[0-9]*' ]]; then
            # Fetch the full list for normal Zsh prefix-based filtering
            while IFS= read -r line; do
                local id="${line%% *}"
                local description="${line#* }"
                insertions+=("$id")
                descriptions+=("${id} ${description}")
            done < <(jt -q "$query_param")

            # Use compadd to display results with the following flags:
            # -Q: Suppresses quoting of completions with special characters. The returned IDs don't need to be quoted.
            # -V jt_issues: Groups completions under the label "jt_issues". This seems to allow
            #    zstyle ':completion::complete:jt::' sort false to work correctly and return the list in the order
            #    that jt -q parents returns them.
            # -d descriptions: Provides descriptions alongside each completion item. Length of insertions and descriptions
            #    must match.
            # -l: Forces a single-column list display regardless of terminal width or number of items
            if compadd -Q -V jt_issues -d descriptions -l -- "${insertions[@]}"; then
                has_matches=1
            fi
        else
            # Perform a text search by appending the search term with a comma.
            query_param+=",$search_term"
            while IFS= read -r line; do
                local id="${line%% *}"
                local description="${line#* }"
                insertions+=("$id")
                descriptions+=("${id} ${description}")
            done < <(jt -q "$query_param")

            # Use compadd to display results. The same flags as above but with and
            # important difference. The -U flag:
            # -U: Ensures Zsh doesn't further filter returned values.
            #    This would filter out the results since the search string would not match the returned IDs.
            if compadd -Q -U -V jt_issues -d descriptions -l-- "${insertions[@]}"; then
                has_matches=1
            fi
        fi

        # Show "No issues found" message if no matches were added
        ((!has_matches)) && compadd -x 'No issues found'
        return 0
        ;;
    query_completion)
        # Define completion options for the -q/--query flag
        local -a query_options
        query_options=("parents" "epics" "initiatives" "tasks" "bugs")

        # Check if the user has entered a comma
        if [[ "$words[CURRENT]" == *,* ]]; then
            # After a comma, allow free text input (no specific completion)
            compadd -U "$words[CURRENT]"
        else
            local current_word="$words[CURRENT]"
            # Filter options based on the current input
            local -a filtered_options
            for opt in "${query_options[@]}"; do
                if [[ -z "$current_word" || "$opt" = ${current_word}* ]]; then
                    filtered_options+=("$opt")
                fi
            done

            # Provide filtered options with no space after completion
            # -S '': Adds an empty string suffix to avoid adding a space after the completion since we support ,<text>
            compadd -Q -U -S '' -- "${filtered_options[@]}"
        fi
        return 0
        ;;
    esac
}

# Register the _jt_completions function for the jt command in Zsh
compdef _jt_completions jt
