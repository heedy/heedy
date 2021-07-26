# The Plugin Template

The previous tutorial saw the development of a simple plugin directly within a heedy database. This works fine for an introduction, but quickly becomes tedious when the plugin becomes more complex, because:
- Any non-trivial frontend code requires a compilation step - this is expecially true if using Vue templates. Furthermore, minification and pre-compression of assets can make the plugin's frontend more efficient, and can be done as part of a build process.
- If including screenshots in the README of your plugin, they won't render in heedy, since the plugin's folder is not exposed to the frontend - a build step is needed to embed images in the README file!
- Debugging and releasing plugins is much easier with an explicit test database and dist folder, which are separate from your plugin's code.

With these benefits in mind, we will now introduce the heedy plugin template: [github.com/heedy/heedy-template-plugin](https://github.com/heedy/heedy-template-plugin). This template sets up a standard structure for building heedy plugins, and is the recommended starting point for people who want to extend heedy.

## Starting with the Template

To begin, we download the most recent release of the template plugin, and extract it.


