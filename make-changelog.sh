git log --pretty="* %s" --no-merges tags/$@..HEAD | grep -E "^\* feat:|^\* fix:" > changelog.md