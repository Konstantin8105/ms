.PHONY: clean
clean:
	@echo "silent remove templorary files"
	find . -name  "*.new"  -type f -delete
	find . -name ".*.new"  -type f -delete
	find . -name  "*.new.png"  -type f -delete
	find . -name ".*.new.png"  -type f -delete
	find . -name  "*.out"  -type f -delete
	find . -name  "*.test" -type f -delete