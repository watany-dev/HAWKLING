# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands
- Build: `npm run build` 
- Lint: `npm run lint`
- Test all: `npm run test`
- Test single: `npm run test -- -t "test name"`

## Code Style Guidelines
- Follow consistent 2-space indentation
- Use camelCase for variables and functions, PascalCase for classes and components
- Prefer const over let, avoid var
- Use TypeScript with explicit types for function parameters and returns
- Sort imports alphabetically: built-ins, third-party, local
- Handle errors explicitly with try/catch; avoid silent failures
- Document public APIs with JSDoc comments
- Use async/await over raw Promises
- Keep functions small and focused on a single responsibility