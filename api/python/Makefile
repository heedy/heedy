.PHONY: dist publish test

dist:
	rm -rf ./dist
	python setup.py sdist bdist_wheel

publish: dist
	twine upload dist/*

test:
	pytest --cov=heedy --timeout=50
