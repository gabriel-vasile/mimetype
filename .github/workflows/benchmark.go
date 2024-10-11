name: Run benchmarks
on:
  pull_request:
    branches: [master]

permissions:
  contents: read

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - name: Install Go
      uses: actions/setup-go@v5.0.2
      with:
        go-version-file: 'go.mod'
    - run: go install golang.org/x/perf/cmd/benchstat@latest
    // Base for comparison is master branch.
    - name: Checkout code
      uses: actions/checkout@v4.1.7
      with:
        ref: master
    - run: go test -run=none -bench=. --count=7 > /tmp/prev &

    - name: Checkout code
      uses: actions/checkout@v4.1.7
    - run: go test -run=none -bench=. --count=7 > /tmp/curr &

    // Wait for both benchmarks to complete before comparing.
    - run: wait
    - run: RESULT="$(benchstat /tmp/prev /tmp/curr)"
    - uses: actions/github-script@v7
        with:
            script: |
                github.rest.issues.createComment({
                    issue_number: context.issue.number,
                    owner: context.repo.owner,
                    repo: context.repo.repo,
                    body: $RESULT
                })
