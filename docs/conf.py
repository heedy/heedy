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
import subprocess
import os
import sys

sys.path.insert(0, os.path.abspath("../api/python"))


# -- Project information -----------------------------------------------------

project = "Heedy"
copyright = "2022, Heedy Contributors"
author = "Heedy Contributors"

# The full version, including alpha/beta/rc tags

release = "%s-git.%s" % (
    open("../VERSION", "r").read().strip(),
    subprocess.run("git rev-list --count HEAD".split(), capture_output=True)
    .stdout.decode()
    .strip(),
)


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "myst_nb",
    "sphinx.ext.autodoc",
    "sphinx.ext.todo",
    "sphinx.ext.mathjax",
    "sphinx.ext.intersphinx",
    "sphinx.ext.viewcode",
    "sphinx.ext.napoleon",
    "sphinx_autodoc_typehints",
    "sphinx_copybutton",
    "sphinx_js",
    "sphinx_inline_tabs",
]

intersphinx_mapping = {
    "aiohttp": ("https://docs.aiohttp.org/en/stable/", None),
}

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = [
    "_build",
    "Thumbs.db",
    ".DS_Store",
    "analysis/pipescript/index_start.md",
]

js_source_path = "../"
jsdoc_config_path = "_jsdoc.json"

# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#

html_theme = "sphinx_book_theme"
html_title = "Heedy"
html_logo = "logo.png"

html_theme_options = {
    "repository_url": "https://github.com/heedy/heedy",
    "repository_branch": "master",
    "use_repository_button": True,
    "use_edit_page_button": True,
    "use_issues_button": True,
    "use_fullscreen_button": False,
    "use_download_button": False,
    "path_to_docs": "docs/",
}

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ["_static"]

html_css_files = ["css/custom.css"]
