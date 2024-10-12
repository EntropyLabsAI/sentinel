cd el
python -m build 

# Upload to testpypi
python3 -m twine upload --repository testpypi dist/*   
python3 -m pip install --index-url https://test.pypi.org/simple/ --no-deps entropy-labs


# Test installation - change the version
pip install --no-cache-dir --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple entropy-labs==0.1.0


Check that everything work as expected


If yes, upload to pypi - make sure to change the version if needed
python -m build 
python3 -m twine upload dist/*


# Test upgrade
pip install entropy-labs --upgrade
