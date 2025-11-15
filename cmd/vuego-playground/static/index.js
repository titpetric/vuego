let examples = {};
let debounceTimer = null;
let isRendering = false;
let isEmbedFS = true; // Will be set by server

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

// Rebuild the example button list from the loaded examples
function refreshExampleButtons() {
    const exampleList = [];
    
    // Build the example list from the examples object
    for (const name in examples) {
        const depth = name.split('/').length - 1;
        const label = name.split('/').pop();
        const isNested = depth > 0;
        const buttonClass = isNested ? 'example-nested' : 'example-root';
        
        exampleList.push({
            name: name,
            label: label,
            depth: depth,
            isNested: isNested,
            buttonClass: buttonClass
        });
    }
    
    // Sort by depth first (root first), then alphabetically
    exampleList.sort((a, b) => {
        if (a.depth !== b.depth) {
            return a.depth - b.depth;
        }
        return a.name.localeCompare(b.name);
    });
    
    // Rebuild Pages group
    const pagesGroup = document.querySelector('.example-group:first-child .button-group');
    pagesGroup.innerHTML = '';
    exampleList.forEach(ex => {
        if (!ex.isNested) {
            const button = document.createElement('button');
            button.className = ex.buttonClass;
            button.textContent = ex.label;
            button.onclick = () => loadExample(ex.name);
            pagesGroup.appendChild(button);
        }
    });
    
    // Rebuild Components group
    const componentsGroup = document.querySelector('.example-group:last-child .button-group');
    componentsGroup.innerHTML = '';
    exampleList.forEach(ex => {
        if (ex.isNested) {
            const button = document.createElement('button');
            button.className = ex.buttonClass;
            button.textContent = ex.label;
            button.onclick = () => loadExample(ex.name);
            componentsGroup.appendChild(button);
        }
    });
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

// Load and display the cheatsheet footer
async function loadCheatsheet() {
    try {
        const response = await fetch('/api/cheatsheet');
        const result = await response.json();
        if (result.content) {
            document.getElementById('cheatsheetContent').innerHTML = result.content;
        }
    } catch (e) {
        console.error('Failed to load cheatsheet:', e);
    }
}

// Show/hide create modal
function showCreateDialog() {
    document.getElementById('createModal').style.display = 'block';
    document.getElementById('fileName').focus();
}

function closeCreateDialog() {
    document.getElementById('createModal').style.display = 'none';
    document.getElementById('fileName').value = '';
}

// Create a new file (page or component)
async function createFile() {
    const fileName = document.getElementById('fileName').value.trim();
    const fileType = document.querySelector('input[name="fileType"]:checked').value;
    
    if (!fileName) {
        const error = document.getElementById('error');
        error.textContent = 'Please enter a file name';
        error.style.display = 'block';
        return;
    }
    
    try {
        const response = await fetch('/api/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: fileName,
                type: fileType
            })
        });
        
        const result = await response.json();
        if (result.error) {
            const error = document.getElementById('error');
            error.textContent = `Error: ${result.error}`;
            error.style.display = 'block';
        } else {
            closeCreateDialog();
            // Reload examples to pick up the new file
            await loadExamples();
            // Refresh the button list to show the new template
            refreshExampleButtons();
            // Load the newly created file directly
            if (result.name && examples[result.name]) {
                const example = examples[result.name];
                document.getElementById('template').value = example.template;
                document.getElementById('data').value = example.data;
                render();
            }
        }
    } catch (e) {
        const error = document.getElementById('error');
        error.textContent = `Failed to create file: ${e.message}`;
        error.style.display = 'block';
    }
}

// Save current template and data to files
async function save() {
    const template = document.getElementById('template').value;
    const data = document.getElementById('data').value;
    
    try {
        const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                template: template,
                data: data
            })
        });
        
        const result = await response.json();
        if (result.error) {
            alert(`Error: ${result.error}`);
        } else {
            alert('Files saved successfully');
        }
    } catch (e) {
        alert(`Failed to save: ${e.message}`);
    }
}

// Close modal when clicking outside of it
window.onclick = function(event) {
    const modal = document.getElementById('createModal');
    if (event.target === modal) {
        modal.style.display = 'none';
    }
};

// Auto-render on load
window.addEventListener('load', async () => {
    // Load examples from server
    await loadExamples();
    
    // Load cheatsheet
    await loadCheatsheet();
    
    // Check if using embedded FS and disable create/save
    try {
        const response = await fetch('/api/status');
        const result = await response.json();
        isEmbedFS = result.isEmbedFS ?? true;
        
        document.getElementById('saveBtn').disabled = isEmbedFS;
        document.getElementById('createBtn').disabled = isEmbedFS;
        
        if (isEmbedFS) {
            document.getElementById('saveBtn').title = 'Save is disabled when using embedded filesystem';
            document.getElementById('createBtn').title = 'Create is disabled when using embedded filesystem';
        }
    } catch (e) {
        console.error('Failed to check server status:', e);
    }
    
    // Render with server-provided initial template and data
    setTimeout(render, 100);
});
