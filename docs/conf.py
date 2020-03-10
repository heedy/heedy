# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
from recommonmark.transform import AutoStructify
import subprocess
import os
import sys
sys.path.insert(0, os.path.abspath('../api/python'))


# -- Project information -----------------------------------------------------

project = 'Heedy'
copyright = '2020, Heedy Contributors'
author = 'Heedy Contributors'

# The full version, including alpha/beta/rc tags

release = "%s-git.%s" % (open('../VERSION', 'r').read().strip(), subprocess.run(
    'git rev-list --count HEAD'.split(), capture_output=True).stdout.decode().strip())


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "recommonmark",
    "sphinx.ext.autodoc",
    "sphinx.ext.todo",
    "sphinx.ext.mathjax",
    "sphinx.ext.intersphinx",
    "sphinx.ext.viewcode",
    "sphinx.ext.napoleon",
    "sphinx_autodoc_typehints",
]

intersphinx_mapping = {
    "aiohttp": ("https://docs.aiohttp.org/en/stable/", None),
}

# Add any paths that contain templates here, relative to this directory.
templates_path = ['_templates']

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store']


# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = 'alabaster'

html_theme_options = {
    'logo': 'logo.png',
    'logo_name': True,
    'logo_text_align': "center",
    'fixed_sidebar': True,
    'github_user': 'heedy',
    'github_repo': 'heedy',
    'github_banner': True,
    'github_button': False,
    'show_powered_by': False,
}

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ['_static']


# https://recommonmark.readthedocs.io/en/latest/#autostructify

def setup(app):
    app.add_transform(AutoStructify)
