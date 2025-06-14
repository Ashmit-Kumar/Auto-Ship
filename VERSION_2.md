**Feature: Intelligent Entrypoint Detection (Version 2 Enhancement)**

**Overview:**
AutoShip Version 2 introduces an enhanced entrypoint detection mechanism for one-click application deployment. This upgrade significantly improves detection accuracy while minimizing performance overhead, enabling smarter and more reliable hosting for user-submitted repositories.

---

**Problem Statement:**
In Version 1, environment detection was based on the presence of specific filenames like `main.go`, `app.py`, or `package.json`. However, real-world projects often use custom file names or nested directories. This limitation leads to false negatives and limits AutoShip's flexibility.

---

**Solution in Version 2:**
A new hybrid detection algorithm has been introduced. It combines file name heuristics with lightweight content inspection to accurately identify the correct entrypoint file for a given programming environment.

---

**Detection Process:**

1. **Initial Heuristic Scan (Fast Check):**

   * Search for common entrypoint filenames:

     * Go: `main.go`
     * Python: `app.py`, `main.py`, `server.py`
     * Node.js: `index.js`, `server.js`, `main.js`
   * If a matching file is found, stop early and use it as the entrypoint.

2. **Optimized Directory Traversal:**

   * Use `filepath.WalkDir()` to walk the directory tree.
   * Ignore unnecessary folders such as:

     * `node_modules/`, `.git/`, `venv/`, `__pycache__/`, `build/`, `dist/`

3. **Extension Filtering:**

   * Only scan files with extensions: `.go`, `.py`, `.js`, `.ts`, `.jsx`, `.tsx`

4. **Lightweight Content Scan (if no match):**

   * For each candidate file, read only the first 2-4 KB.
   * Check for environment-specific signatures:

     * Go: `package main`, `func main()`
     * Python: `if __name__ == "__main__"`
     * Node.js: `express()`, `createServer()`

5. **Fail Fast Behavior:**

   * Exit scanning as soon as a valid signature is detected.

---

**Performance Impact:**

* Designed to complete within 200ms for typical repositories.
* Efficient even for moderately large codebases (up to 1000 files).

---

**Benefits:**

* Increased accuracy in environment detection
* Supports flexible project structures
* Avoids false negatives due to unconventional file naming
* Maintains fast performance

---

**Future Considerations:**

* Support for `Procfile`, `start.sh`, and Dockerfile parsing
* Use of language-specific AST parsers for more precise detection
* Optional developer hint via `autoship.json`

---

**Status:**

* Scheduled for inclusion in Version 2 release
* Implementation module: `services/entrypoint.go`
* Unit tests and benchmarks pending

---

**Conclusion:**
This feature enables AutoShip to better adapt to real-world project diversity, improving the reliability of automatic containerization and deployment across all supported environments.
