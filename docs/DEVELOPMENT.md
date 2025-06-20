# Development Workflow Guide

## Required Development Process

This document outlines the mandatory workflow for making changes to the Sleeper MCP Server.

## 1. Before Making Changes

1. **Understand the codebase**: Read existing documentation and code
2. **Create a plan**: Use TodoWrite tool to track what needs to be done
3. **Write tests first** (when applicable): Follow TDD principles

## 2. Development Cycle

### Make Changes
- Implement features/fixes following existing patterns
- Add appropriate logging and error handling
- Update documentation as you go

### Validate Changes
**MANDATORY**: After every body of changes, run:

```bash
make test
```

This runs both unit and integration tests and **MUST PASS** before proceeding.

### Test Failure Resolution
If tests fail:
1. **Do not proceed** with additional changes
2. Fix the failing tests first
3. Re-run `make test` until all tests pass
4. Only then continue with development

## 3. Types of Changes Requiring Full Test Validation

- ✅ New tool implementations
- ✅ Modifications to existing tools
- ✅ Changes to data models or interfaces
- ✅ Refactoring of core logic
- ✅ Bug fixes
- ✅ Configuration changes
- ✅ Dependency updates
- ✅ Documentation that affects code behavior

## 4. Commit Guidelines

### Before Committing
1. ✅ All tests pass (`make test`)
2. ✅ Code builds without warnings
3. ✅ Documentation is updated
4. ✅ TodoWrite used to track completed work

### Commit Message Format
```
<type>: <description>

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 5. Testing Commands Reference

### Validation Testing
```bash
# Run all tests (required after changes)
make test

# Run only unit tests
go test ./...

# Run only integration tests  
go test -tags=integration ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestSpecificFunction ./internal/handlers
```

### MCP Integration Testing
Test tools through the actual MCP protocol:

```bash
# Show MCP testing instructions
make test-mcp
```

#### Setup for MCP Testing
1. **Configure Claude Desktop**: Update your Claude Desktop config with the server details (see `configs/claude_desktop_config.json`)
2. **Test naturally**: Use conversational commands like:
   - "Get standings for league YOUR_LEAGUE_ID with custom tiebreakers: when teams have the same wins, use head-to-head record first, then points for"
   - "Show me all users in league YOUR_LEAGUE_ID"
   - "Get matchups for week 1 in league YOUR_LEAGUE_ID"

This approach tests both MCP protocol integration and tool functionality as they would be used in practice.

## 6. Common Development Patterns

### Adding New Tools
1. Define args struct in handlers package
2. Implement tool definition method
3. Implement tool handler method
4. Add tool to MCP server registration
5. Write comprehensive tests
6. **Run `make test`**
7. Update documentation

### Modifying Existing Tools
1. Update implementation
2. Update/add tests for new behavior
3. **Run `make test`**
4. Update documentation
5. Update any affected examples

### Bug Fixes
1. Write a test that reproduces the bug
2. Verify test fails
3. Fix the bug
4. Verify test now passes
5. **Run `make test`**
6. Document the fix

## 7. Emergency Procedures

### If Tests Break During Development
1. **STOP** adding new features
2. Revert changes to last working state if necessary
3. Fix tests one by one
4. Ensure `make test` passes before continuing

### If Integration Tests Fail
1. Check if Sleeper API is accessible
2. Verify test data is still valid
3. Check rate limiting issues
4. Fix tests or mark as skip if external service issue

## 8. Quality Gates

All changes must pass these gates:

- [ ] `make test` passes completely
- [ ] Code builds without errors/warnings
- [ ] New functionality has tests
- [ ] Documentation is updated
- [ ] TodoWrite tracking is complete
- [ ] Changes follow existing code patterns

## 9. League Configuration

### League Settings File
- League-specific configurations are stored in `configs/league_settings.json`
- Custom standings calculations can be pre-configured per league
- The server automatically applies league-specific rules when processing requests
- Works with any MCP-compatible AI agent (Claude, ChatGPT, etc.)

### Configuration Structure
Each league can have:
- Custom tiebreaker instructions in natural language
- Specific tiebreaker ordering
- Additional notes about league rules
- Enable/disable flags for custom calculations

## 10. Best Practices

### Testing
- Write tests for both success and error cases
- Use descriptive test names
- Mock external dependencies in unit tests
- Test edge cases and boundary conditions

### Code Quality
- Follow Go idioms and conventions
- Add proper error handling
- Include appropriate logging
- Use existing utility functions when possible

### API Integration Principles
- **Use authoritative data sources**: Always prefer explicit API fields over inference
- **Leverage Sleeper's structured data**: Use bracket endpoints (`/winners_bracket`, `/losers_bracket`) for definitive playoff results
- **Avoid assumptions**: Don't infer game types (championship vs third place) from patterns - use Sleeper's explicit bracket data
- **Trust the source**: Sleeper's API provides authoritative bracket structure, matchup results, and playoff positioning

### Documentation
- Update tool specifications for API changes
- Add examples for complex features
- Document any breaking changes
- Keep README.md current
- Update `league_settings.json` when league-specific behavior changes

## 11. Troubleshooting

### Common Test Failures
- **Import cycle**: Check for circular dependencies
- **Missing mocks**: Ensure MockSleeperClient implements all required methods
- **API changes**: Update models if Sleeper API has changed
- **Timeout issues**: Increase test timeouts for slow operations

### Getting Help
1. Check existing documentation in `/docs`
2. Review similar implementations in codebase
3. Check Git history for context on changes
4. Review test output for specific failure details

Remember: **The test suite is your safety net. Never skip `make test` after changes!**