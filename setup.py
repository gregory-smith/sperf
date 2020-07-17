"""setup script for sperf"""

from setuptools import setup
#setuptools import is needed on 3.5
#pylint: disable=unused-import
import setuptools

setup(
    name='sperf',
    version='0.6.2',
    description='Diagnostic utility for DSE and Cassandra',
    url='https://www.github.com/DataStax-Toolkit/sperf',
    app=["scripts/sperf"],
    setup_requires=['PyInstaller', 'pytest'],
    package_dir={'': 'pysper'},
)