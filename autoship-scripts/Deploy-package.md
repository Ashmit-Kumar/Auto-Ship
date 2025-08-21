Got it ✅ — let’s focus **only on the GitHub Pages repo part**, with your repo name as **`autoship-deploy`**.

---

## 📦 Using GitHub Pages as Yum Repo (for `autoship-deploy`)

### 1. Create a repo on GitHub

* Name it: **`autoship-deploy`**
* Enable **GitHub Pages** (from repo settings → Pages → set branch `gh-pages`)

Your repo will be served at:

```
https://<your-username>.github.io/autoship-deploy/
```

---

### 2. Build RPM and repo metadata

Locally (or in CI), after you create your `.rpm` for `autoship-deploy`:

```bash
mkdir repo
mv autoship-deploy-1.0.0-1.x86_64.rpm repo/
createrepo repo/
```

Now `repo/` looks like:

```
repo/
 ├── autoship-deploy-1.0.0-1.x86_64.rpm
 └── repodata/
```

---

### 3. Push to `gh-pages` branch

* Commit everything inside `repo/` to the **`gh-pages`** branch of your repo
* GitHub Pages will serve it at:

  ```
  https://<your-username>.github.io/autoship-deploy/
  ```

---

### 4. Add Yum repo file on EC2

On any EC2 instance, create `/etc/yum.repos.d/autoship.repo`:

```ini
[autoship-deploy]
name=Autoship Deploy Repo
baseurl=https://<your-username>.github.io/autoship-deploy/
enabled=1
gpgcheck=0
```

---

### 5. Install anywhere

Now you can do:

```bash
sudo yum install autoship-deploy
```

And later update:

```bash
sudo yum update autoship-deploy
```

---

✅ This makes **`autoship-deploy` globally installable** from GitHub Pages, no manual copy-paste to `/opt/` anymore.

