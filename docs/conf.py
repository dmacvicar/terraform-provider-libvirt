# -*- coding: utf-8 -*-
import subprocess
import os
import sys
sys.path.insert(0, os.path.abspath('.'))
from recommonmark.transform import AutoStructify

project = 'terraform-provider-libvirt'
copyright = '2019, Duncan Mac-Vicar P. & contributors'
author = 'Duncan Mac-Vicar P.'

version = subprocess.check_output(('git', 'describe', '--tags')).decode("utf-8")

source_suffix = {
    '.html.markdown': 'markdown',
}

templates_path = ['_templates']
exclude_patterns = ["_build"]
html_theme = 'sphinx_rtd_theme'
extensions = ['removefrontmatter','recommonmark']

html_theme_options = {
    'collapse_navigation': False,
}

master_doc = 'toc'

def setup(app):
    app.add_config_value('recommonmark_config', {
        'enable_eval_rst': True,
        'auto_toc_tree_section': 'Resources',
        'auto_toc_maxdepth': 3,
        'commonmark_suffixes': ['.html.markdown'],
    }, 'env')
    app.add_transform(AutoStructify)


