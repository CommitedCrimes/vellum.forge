---
title: "Bluemonday Security Test"
date: 2024-01-25T10:00:00Z
tags: ["security", "test", "bluemonday"]
description: "Testing HTML sanitization with Bluemonday"
draft: false
slug: "bluemonday-test"
---

# Bluemonday Security Test

This post tests HTML sanitization with Bluemonday.

## Safe HTML (should be preserved)

This is **bold text** and this is *italic text*.

- List item 1
- List item 2

```go
package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}
```

## Potentially Dangerous HTML (should be sanitized)

<script>alert('XSS Attack!')</script>

<img src="x" onerror="alert('XSS')">

<a href="javascript:alert('XSS')">Click me</a>

<iframe src="malicious-site.com"></iframe>

## Mermaid Diagram (should be preserved)

```mermaid
graph TD
    A[Start] --> B{Is it safe?}
    B -->|Yes| C[Allow]
    B -->|No| D[Sanitize]
    D --> E[Clean Output]
```

## Conclusion

Bluemonday should sanitize dangerous HTML while preserving safe formatting and code blocks.
