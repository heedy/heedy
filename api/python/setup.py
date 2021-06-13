import pathlib
import setuptools

# The directory containing this file
HERE = pathlib.Path(__file__).parent

# The text of the README file
README = (HERE / "README.md").read_text()

# This call to setup() does all the work
setuptools.setup(
    name="heedy",
    version="0.1.2",
    description="A Python library for interfacing with Heedy",
    long_description=README,
    long_description_content_type="text/markdown",
    url="https://github.com/heedy/heedy",
    author="Heedy Contributors",
    license="apache2",
    classifiers=[
        "Intended Audience :: Developers",
        "Intended Audience :: Science/Research",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
    ],
    packages=setuptools.find_packages(),
    include_package_data=True,
    python_requires=">=3.7.0",
    install_requires=["aiohttp", "dateparser", "requests"],
    extras_require={"dataframes": ["pandas"]},
)
