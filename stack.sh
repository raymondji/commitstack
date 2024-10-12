#!/bin/bash
GS_BASE_BRANCH=${GS_BASE_BRANCH:-main}
GS_ENABLE_COLOR_OUTPUT=${GS_ENABLE_COLOR_OUTPUT:-yes} # no to disable
GS_ENABLE_GITLAB_EXTENSION=${GS_ENABLE_GITLAB_EXTENSION:-no} # yes to enable
GS_ENABLE_GITHUB_EXTENSION=${GS_ENABLE_GITHUB_EXTENSION:-no} # yes to enable

gs() {
    git-stacked "$@"
}

git-stacked() {
    if [ $# -eq 0 ]; then
        echo "Must provide subcommand"
        git-stacked-help
        return 1
    fi

    COMMAND=$1
    shift

    if [ "$GS_ENABLE_GITLAB_EXTENSION" = "yes" ]; then
        if ! command -v jq &> /dev/null || ! command -v glab &> /dev/null; then
            return 1
        fi
    elif [ "$GS_ENABLE_GITHUB_EXTENSION" = "yes" ]; then
        if ! command -v jq &> /dev/null || ! command -v gh &> /dev/null; then
            return 1
        fi
    fi

    USE_EXTENSION=none
    REMOTE_URL=$(git remote get-url origin)
    if [[ "$REMOTE_URL" == *"gitlab.com"* ]] && [[ "$GS_ENABLE_GITLAB_EXTENSION" == "yes" ]]; then
        USE_EXTENSION="gitlab"
    elif [[ "$REMOTE_URL" == *"github.com"* ]] && [[ "$GS_ENABLE_GITHUB_EXTENSION" == "yes" ]]; then
        USE_EXTENSION="github"
    fi

    if [ "$COMMAND" = "help" ] || [ "$COMMAND" = "h" ]; then
        git-stacked-help
    elif [ "$COMMAND" = "push-force" ] || [ "$COMMAND" = "pf" ]; then
        if [ $USE_EXTENSION = "gitlab" ]; then
            gitlab-stacked-push-force
        elif [ $USE_EXTENSION = "github" ]; then
            github-stacked-push-force
        else
            git-stacked-push-force
        fi
    elif [ "$COMMAND" = "create" ] || [ "$COMMAND" = "c" ]; then
        git-stacked-create "$@"
    elif [ "$COMMAND" = "pull-rebase" ] || [ "$COMMAND" = "pr" ]; then
        git-stacked-pull-rebase
    elif [ "$COMMAND" = "rebase" ] || [ "$COMMAND" = "r" ]; then
        git-stacked-rebase
    elif [ "$COMMAND" = "branch" ] || [ "$COMMAND" = "b" ]; then
        git-stacked-branch
    elif [ "$COMMAND" = "stack" ] || [ "$COMMAND" = "s" ]; then
        git-stacked-stack
    elif [ "$COMMAND" = "log" ] || [ "$COMMAND" = "l" ]; then
        git-stacked-log
    elif [ "$COMMAND" = "reorder" ] || [ "$COMMAND" = "ro" ]; then
        git-stacked-reorder
    else
        echo "Invalid command"
        echo ""
        git-stacked-help
    fi
}

git-stacked-help() {
    echo 'usage: git-stacked ${subcommand} ...
    alias: gs

subcommands:

create
    alias: c
    create a new branch on top of the current stack

push-force
    alias: pf
    push all branches in the current stack to remote

pull-rebase
    alias: pr
    update the base branch from mainstream, then rebase the current stack onto the base branch

rebase
    alias: r
    start interactive rebase of the current stack against the base branch

log
    alias: l
    git log helper

stack
    alias: s
    list all stacks

branch
    alias: b
    list all branches in the current stack

reorder
    alias: ro
    start interactive rebase to reorder branches in the current stack'
}

git-stacked-create() {
    BRANCH=$1
    git checkout -b $BRANCH
    git commit --allow-empty -m "Start of $BRANCH"
}

# is_top_of_stack returns 0 if the current branch is the tip of a stack
# otherwise 1
is_top_of_stack() {
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    local DESCENDENT_COUNT=$(git branch --contains "$CURRENT_BRANCH" | wc -l)
    if [[ "$DESCENDENT_COUNT" -eq 1 ]]; then
        return 0
    else
        return 1
    fi
}

git-stacked-branch() {
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    BRANCHES=$(git log --pretty='format:%D' $GS_BASE_BRANCH.. --decorate-refs=refs/heads | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "Not in a stack"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        if [ "$BRANCH" = "$CURRENT_BRANCH" ]; then
            if [ "$GS_ENABLE_COLOR_OUTPUT" = "yes" ]; then
                echo -n "* \033[0;32m$BRANCH\033[0m" # green highlight
                if is_top_of_stack; then
                    echo " (top)"
                else
                    echo ""
                fi
            else
                echo -n "* $BRANCH"
                if is_top_of_stack; then
                    echo " (top)"
                else
                    echo ""
                fi
            fi
        else
            echo "  $BRANCH"
        fi
    done
}

git-stacked-stack() {
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    BRANCHES=$(git branch --format='%(refname:short)')
    STACKS=()

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        if [[ "$BRANCH" == "$GS_BASE_BRANCH" ]]; then
            continue
        fi

        DESCENDENT_COUNT=$(git branch --contains "$BRANCH" | wc -l)
        # Branches are always a descendent of themselves, so 1 means there are no other descendents.
        # i.e. this branch is the tip of a stack.
        if [[ "$DESCENDENT_COUNT" -eq 1 ]]; then
            STACKS+=("$BRANCH")
        fi
    done


    CONTAINING_CURRENT=$(git branch --contains "$CURRENT_BRANCH")
    if [[ "$CURRENT_BRANCH" == "$GS_BASE_BRANCH" ]]; then
        CONTAINING_CURRENT=""
    fi

    for STACK in "${STACKS[@]}"; do
        if echo "$CONTAINING_CURRENT" | grep -q "$STACK"; then
            if [ "$GS_ENABLE_COLOR_OUTPUT" = "yes" ]; then
                echo -e "* \033[0;32m$STACK\033[0m" # green highlight
            else
                echo "* $STACK"
            fi
        else
            echo "  $STACK"
        fi
    done
}

git-stacked-log() {
    git log $GS_BASE_BRANCH..
}

git-stacked-push-force() {
    # Reverse so we push from bottom -> top
    BRANCHES=$(git log --pretty='format:%D' $GS_BASE_BRANCH.. --decorate-refs=refs/heads --reverse | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "Not in a stack"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo "branch: $BRANCH"
        echo "----------------------------"
        git push origin "$BRANCH":"$BRANCH" --force
        echo "" # newline
    done
}

gitlab-stacked-push-force() {
    echo "Gitlab extension not implemented yet, falling back to default behaviour."
    echo ""
    git-stacked-push-force
}

github-stacked-push-force() {
    # Reverse so we push from bottom -> top
    BRANCHES=$(git log --pretty='format:%D' $GS_BASE_BRANCH.. --decorate-refs=refs/heads --reverse | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "Not in a stack"
        return 1
    fi

    # First reset the base branch to $GS_BASE_BRANCH for all existing MRs.
    # If the branches have been re-ordered, this prevents unintentional merging.
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo "Prepare branch: $BRANCH"
        echo "----------------------------"
        local PR_EXISTS=$(gh pr list --head "$BRANCH" --json number | jq '. | length')
        if [ "$PR_EXISTS" -gt 0 ]; then
            local PR_NUMBER=$(gh pr list --head "$BRANCH" --json number | jq -r '.[0].number')
            echo "Changing PR target branch to $GS_BASE_BRANCH for PR #$PR_NUMBER..."
            gh pr edit "$PR_NUMBER" --base "$GS_BASE_BRANCH"
        fi
        echo "" # Print a newline for readability
    done

    local PREVIOUS_BRANCH="$GS_BASE_BRANCH"
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo "Push branch: $BRANCH"
        echo "----------------------------"
        git push origin "$BRANCH:$BRANCH" --force
        
        local PR_EXISTS=$(gh pr list --head "$BRANCH" --json number | jq '. | length')
        if [ "$PR_EXISTS" -eq 0 ]; then
            echo "Creating a new PR for branch $BRANCH..."
            gh pr create --base "$PREVIOUS_BRANCH" --head "$BRANCH" --title "PR for $BRANCH" --body "This PR was created automatically."
        else
            if [ "$PREVIOUS_BRANCH" != "$GS_BASE_BRANCH" ]; then
                echo "Changing PR target branch back to $PREVIOUS_BRANCH for PR #$PR_NUMBER..."
                gh pr edit "$PR_NUMBER" --base "$PREVIOUS_BRANCH"
            fi
        fi

        PREVIOUS_BRANCH="$BRANCH"
        echo "" # Print a newline for readability
    done
}

git-stacked-pull-rebase() {
    git checkout $GS_BASE_BRANCH && \
    git pull && \
    git checkout - && \
    git rebase -i $GS_BASE_BRANCH --update-refs
}

git-stacked-rebase() {
    git rebase -i $GS_BASE_BRANCH --update-refs --keep-base
}

git-stacked-reorder() {
    echo "Before continuing, please set the target branch of all open merge requests in this stack to $GS_BASE_BRANCH"
    echo "Done: [Y/n]"
    read input
    if [[ "$input" == "Y" || "$input" == "y" ]]; then
       echo "Proceeding..."
    else
        echo "Exiting..."
        return 1
    fi

    git checkout -b tmp-reorder-branch && \
    git rebase -i $GS_BASE_BRANCH --update-refs --keep-base && \
    git checkout - && \
    git branch -D tmp-reorder-branch
}
