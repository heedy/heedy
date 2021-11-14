.PHONY: dist publish test clean

dist:
	rm -rf ./dist
	python setup.py sdist bdist_wheel

publish: dist
	twine upload dist/*

test:
	./runtests

clean:
	rm -rf ./dist
	rm -rf ./build
	rm -rf ./heedy.egg-info
	find . -type f -name '*.py[co]' -delete -o -type d -name __pycache__ -delete