# Build the package
python -m build

# Upload to TestPyPI - Skip this if you want to directly upload to PyPI
python3 -m twine upload --repository testpypi dist/*
python3 -m pip install --index-url https://test.pypi.org/simple/ asteroid_sdk

# Test installation (update version as needed)
pip install --no-cache-dir --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple asteroid_sdk==0.1.0

# Verify functionality
# If successful, upload to PyPI - Start here if you want to upload to PyPI directly
python -m build
python3 -m twine upload dist/*

# Test upgrade
pip install asteroid_sdk --upgrade

# Local installation verification
pip cache purge
pip-autoremove asteroid_sdk -y

# Install locally
pip install -e ".[dev]"
pip install dist/asteroid_sdk-0.1.0-py3-none-any.whl

# OpenAPI Python client setup
# If using the OpenAPI Python library, follow these steps:
# https://github.com/openapi-generators/openapi-python-client

# Generate or update the client
openapi-python-client generate --path server/openapi.yaml --output-path asteroid_sdk/src/asteroid_sdk/sentinel_api_client --overwrite

# Database setup (manual steps)
psql -h 127.0.0.1 -p 5433 -U root -d sentinel -f schema.sql