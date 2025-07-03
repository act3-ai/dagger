"""Tests for main.py"""

import main


def test_main_output(capfd):
    """Check printout is the same as main.py"""
    main.main()
    out, _err = capfd.readouterr()
    assert out.strip() == "Hello from testapp!"
