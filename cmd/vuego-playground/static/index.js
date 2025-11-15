let examples = {};
let debounceTimer = null;
let isRendering = false;

// Load examples from the server (for getting template/data when a button is clicked)
async function loadExamples() {
    try {
        const response = await fetch('/api/examples');
        const result = await response.json();
        if (result.examples) {
            examples = result.examples;
        }
    } catch (e) {
        console.error('Failed to load examples:', e);
    }
}

function loadExample(name) {
    const example = examples[name];
    if (example) {
        document.getElementById('template').value = example.template;
        document.getElementById('data').value = example.data;
        render();
    }
}

async function render() {
    if (isRendering) return;
    
    const template = document.getElementById('template').value;
    const dataStr = document.getElementById('data').value;
    const spinner = document.getElementById('spinner');
    const status = document.getElementById('status');
    const preview = document.getElementById('preview');
    const error = document.getElementById('error');

    // Parse data JSON first to validate
    let data;
    try {
        data = JSON.parse(dataStr);
    } catch (e) {
        error.textContent = `Invalid JSON data: ${e.message}`;
        error.style.display = 'block';
        preview.style.display = 'none';
        return;
    }

    isRendering = true;
    spinner.style.display = 'block';
    status.textContent = 'Rendering...';
    error.style.display = 'none';
    preview.style.display = 'none';

    try {
        const response = await fetch('/api/render', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                template: template,
                data: dataStr
            })
        });

        const result = await response.json();

        if (result.error) {
            error.textContent = result.error;
            error.style.display = 'block';
            preview.style.display = 'none';
            status.textContent = 'Error';
        } else {
            // Create isolated iframe document with base styling
            const iframeDoc = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            padding: 2rem;
            line-height: 1.6;
            color: #333;
        }
        h1, h2, h3, h4, h5, h6 {
            margin: 1.5em 0 0.5em 0;
            line-height: 1.3;
            font-weight: 600;
            color: #1a1a1a;
        }
        h1:first-child, h2:first-child, h3:first-child {
            margin-top: 0;
        }
        h1 { font-size: 2em; }
        h2 { font-size: 1.5em; }
        h3 { font-size: 1.25em; }
        h4 { font-size: 1.1em; }
        h5 { font-size: 1em; }
        h6 { font-size: 0.9em; }
        p { margin: 0 0 1em 0; }
        ul, ol {
            margin: 0 0 1em 0;
            padding-left: 2em;
        }
        li { margin: 0.25em 0; }
        ul ul, ol ol, ul ol, ol ul { margin: 0.25em 0; }
        strong, b { font-weight: 600; }
        em, i { font-style: italic; }
        u { text-decoration: underline; }
        a {
            color: #007d9c;
            text-decoration: none;
        }
        a:hover { text-decoration: underline; }
        code {
            background: #f5f5f5;
            padding: 0.2em 0.4em;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9em;
        }
        pre {
            background: #f5f5f5;
            padding: 1em;
            border-radius: 4px;
            overflow-x: auto;
            margin: 0 0 1em 0;
        }
        pre code {
            background: none;
            padding: 0;
        }
        blockquote {
            border-left: 4px solid #e9ecef;
            margin: 0 0 1em 0;
            padding: 0.5em 0 0.5em 1em;
            color: #666;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 0 0 1em 0;
        }
        th, td {
            border: 1px solid #e9ecef;
            padding: 0.5em;
            text-align: left;
        }
        th {
            background: #f8f9fa;
            font-weight: 600;
        }
        hr {
            border: none;
            border-top: 1px solid #e9ecef;
            margin: 2em 0;
        }
        input[type="checkbox"] {
            margin-right: 0.5em;
        }
    </style>
</head>
<body>
${result.html}
</body>
</html>`;
            preview.srcdoc = iframeDoc;
            preview.style.display = 'block';
            error.style.display = 'none';
            status.textContent = 'Success';
            setTimeout(() => status.textContent = '', 2000);
        }
    } catch (e) {
        error.textContent = `Network error: ${e.message}`;
        error.style.display = 'block';
        preview.style.display = 'none';
        status.textContent = 'Error';
    } finally {
        isRendering = false;
        spinner.style.display = 'none';
    }
}

function scheduleRender() {
    if (!document.getElementById('autoRefresh').checked) return;
    
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
        render();
    }, 500);
}

// Setup auto-refresh listeners
document.getElementById('template').addEventListener('input', scheduleRender);
document.getElementById('data').addEventListener('input', scheduleRender);

// Keyboard shortcut: Ctrl+Enter or Cmd+Enter to render
document.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        render();
    }
});

// Auto-render on load
window.addEventListener('load', async () => {
    // Load examples from server
    await loadExamples();
    
    // Render with server-provided initial template and data
    setTimeout(render, 100);
});
