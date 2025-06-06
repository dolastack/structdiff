# structdiff

A fast, cross-platform CLI tool to compare structured data files (JSON, YAML, TOML, XML, INI, CSV) and show differences in a human-readable format.

---

## ✨ Features

- 🔍 **Compare structured files:** JSON, YAML, TOML, XML, INI, CSV
- 🔄 **Fuzzy key matching:** Case-insensitive field alignment
- 🧠 **Deep nested diffing:** Handles maps, slices, and deeply nested structures
- 🎯 **Filter by path:** Only show diffs under specific keys
- 🚫 **Exit on mismatch:** Integrate with CI/CD pipelines (`--check`)
- 🖌️ **Colorized output:** Easy-to-read terminal output
- 🌐 **Remote URLs:** Compare files from HTTP(S) endpoints
- 🔐 **Authentication support:**
  - Basic Auth
  - Bearer Tokens
  - OAuth2 Client Credentials
  - AWS Sigv4 Signing
  - AWS IAM Role Assumption (STS)
  - SSO via Device Code Flow
- ⚙️ **Configurable via `.structdiff.yaml`**
- 🐳 **Docker image available**
- 📦 **Reusable Go library for integration into other tools**

---

## 🧩 Basic Usage

### Sample Files

**file1.yaml**
```yaml
User:
  Name: Alice
  Age: 30
```

**file2.yaml**
```yaml
user:
  name: Bob
  age: 30
  email: alice@example.com
```

### Compare Two Files

```sh
structdiff compare file1.yaml file2.yaml
```

#### Example Output

```bash
Found 2 differences
.user.name    Value mismatch: 'Alice' vs 'Bob'
.user.email   Only in right: 'alice@example.com'
```

---

## 🏳️ Flags

| Flag                | Description                                             |
|---------------------|--------------------------------------------------------|
| `-f`, `--filter`    | Only show diffs under the specified key path           |
| `--check`           | Exit with non-zero code if differences are found       |

---

## 🐳 Docker Usage

```sh
docker run --rm -v $PWD:/data dolastack/structdiff compare /data/file1.yaml /data/file2.yaml
```

---

## ⚙️ Configuration

You can configure structdiff using a `.structdiff.yaml` file in your project directory.

---

## 🤝 Contributing

Contributions are welcome! Please see `CONTRIBUTING.md` for guidelines.

---

## 📄 License

MIT License. See `LICENSE` for details.