# Deployment Guide

The browser version of Podhnologic is a static web application that can be deployed to any static hosting service.

## Quick Deploy Options

### 1. GitHub Pages (Free)

```bash
# From repository root
git checkout -b gh-pages
git add browser/
git commit -m "Deploy browser version"
git subtree push --prefix browser origin gh-pages
```

Then enable GitHub Pages in repository settings, pointing to the `gh-pages` branch.

**URL**: `https://yourusername.github.io/podhnologic/`

### 2. Netlify (Free)

1. Create account at [netlify.com](https://netlify.com)
2. Click "Add new site" → "Import an existing project"
3. Connect your repository
4. Set build settings:
   - **Base directory**: `browser`
   - **Build command**: (leave empty)
   - **Publish directory**: `.` (current directory)
5. Click "Deploy site"

**Drag & Drop Alternative:**
- Just drag the `browser/` folder into Netlify's drop zone

### 3. Vercel (Free)

```bash
# Install Vercel CLI
npm i -g vercel

# From browser directory
cd browser
vercel
```

Follow the prompts to deploy.

### 4. Cloudflare Pages (Free)

1. Create account at [pages.cloudflare.com](https://pages.cloudflare.com)
2. Click "Create a project" → "Connect to Git"
3. Select your repository
4. Set build settings:
   - **Build command**: (leave empty)
   - **Build output directory**: `browser`
5. Click "Save and Deploy"

### 5. Local Server (Development)

```bash
cd browser
npm install
npm start
```

Open http://localhost:8080

### 6. Simple HTTP Server (Any language)

Python:
```bash
cd browser
python3 -m http.server 8080
```

Node.js:
```bash
cd browser
npx http-server -p 8080 -c-1 --cors
```

PHP:
```bash
cd browser
php -S localhost:8080
```

## CORS Requirements

The application loads FFmpeg.wasm from a CDN (unpkg.com). Most hosting services handle CORS automatically, but if you encounter issues:

1. Ensure your server sends proper CORS headers
2. Use `--cors` flag if using http-server
3. Consider hosting FFmpeg.wasm locally (see Advanced section)

## Advanced: Self-Host FFmpeg.wasm

To make the app 100% self-contained:

1. Download FFmpeg.wasm files:
```bash
cd browser
mkdir -p lib
curl -L https://unpkg.com/@ffmpeg/ffmpeg@0.12.10/dist/esm/index.js -o lib/ffmpeg.js
curl -L https://unpkg.com/@ffmpeg/util@0.12.1/dist/esm/index.js -o lib/util.js
curl -L https://unpkg.com/@ffmpeg/core@0.12.6/dist/esm/ffmpeg-core.js -o lib/ffmpeg-core.js
curl -L https://unpkg.com/@ffmpeg/core@0.12.6/dist/esm/ffmpeg-core.wasm -o lib/ffmpeg-core.wasm
```

2. Update imports in `app.js`:
```javascript
// Change from:
import { FFmpeg } from 'https://unpkg.com/@ffmpeg/ffmpeg@0.12.10/dist/esm/index.js';

// To:
import { FFmpeg } from './lib/ffmpeg.js';
```

3. Update core loading paths in `app.js`:
```javascript
// Change from:
const baseURL = 'https://unpkg.com/@ffmpeg/core@0.12.6/dist/esm';

// To:
const baseURL = './lib';
```

## Custom Domain

After deploying to any service above:

1. Go to your hosting service's settings
2. Add custom domain
3. Update DNS records (usually a CNAME)
4. Wait for SSL certificate to be issued (automatic)

## Security Headers (Optional)

Add these headers for enhanced security:

```
Content-Security-Policy: default-src 'self' 'unsafe-inline' 'unsafe-eval' blob: data: https://unpkg.com;
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
```

Most hosting services allow custom headers in their configuration.

## Performance Optimization

### Enable Compression

Most hosting services enable gzip/brotli automatically. Verify with:
```bash
curl -H "Accept-Encoding: gzip" -I https://your-domain.com/app.js
```

### Cache Headers

Set long cache times for static assets:
- `index.html`: no-cache
- `app.js`: max-age=31536000 (1 year)
- FFmpeg files: max-age=31536000 (1 year)

Example Netlify `_headers` file:
```
/index.html
  Cache-Control: no-cache

/app.js
  Cache-Control: public, max-age=31536000, immutable

/lib/*
  Cache-Control: public, max-age=31536000, immutable
```

## Monitoring

Most hosting services provide analytics:
- Netlify: Built-in analytics
- Vercel: Built-in analytics
- Cloudflare: Built-in analytics
- Custom: Add privacy-friendly analytics like Plausible or Fathom

## Troubleshooting

### Issue: "SharedArrayBuffer is not defined"

**Solution**: This error occurs if CORS headers aren't set correctly. FFmpeg.wasm requires:
```
Cross-Origin-Embedder-Policy: require-corp
Cross-Origin-Opener-Policy: same-origin
```

Add these headers in your hosting service configuration.

### Issue: Module loading errors

**Solution**: Ensure your server sends correct MIME types:
- `.js` → `text/javascript`
- `.wasm` → `application/wasm`

### Issue: File System Access API not working

**Solution**: This API requires:
- Chromium-based browser (Chrome, Edge, Opera)
- Secure context (HTTPS or localhost)

## Cost Estimates

All recommended services offer generous free tiers:

- **GitHub Pages**: Free (public repos), 1GB bandwidth/month
- **Netlify**: Free tier (100GB bandwidth/month)
- **Vercel**: Free tier (100GB bandwidth/month)
- **Cloudflare Pages**: Free tier (unlimited bandwidth!)

For a low-traffic personal project, all these services will remain free indefinitely.

## Updates

To update your deployed version:

```bash
# Make changes to browser/ files
git add browser/
git commit -m "Update browser version"
git push

# Most services will auto-deploy on push
# For GitHub Pages, push to gh-pages branch
```

## Questions?

See main README or open an issue on GitHub.
