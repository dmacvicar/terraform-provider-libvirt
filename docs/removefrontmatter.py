import re

# Remove the frint matter (YAML) from markdown files
def source_read_handler(app, docname, source):
    source[0] = re.sub(r'^---\n(.+\n)*---\n', '', source[0])

def setup(app):
    app.connect('source-read', source_read_handler)
