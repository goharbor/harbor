import setuptools

try:
    import multiprocessing  # noqa
except ImportError:
    pass

setuptools.setup(setup_requires=['pbr>=1.8'], pbr=True)
