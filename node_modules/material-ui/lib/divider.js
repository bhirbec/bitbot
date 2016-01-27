'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _react = require('react');

var _react2 = _interopRequireDefault(_react);

var _muiThemeable = require('./muiThemeable');

var _muiThemeable2 = _interopRequireDefault(_muiThemeable);

var _styles = require('./utils/styles');

var _styles2 = _interopRequireDefault(_styles);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var propTypes = {
  /**
   * CSS class that will be added to the divider's root element
   */
  className: _react2.default.PropTypes.string,

  /**
   * If true, the divider will be indented 72px
   */
  inset: _react2.default.PropTypes.bool,

  /**
   * Override the inline-styles of the list divider's root element
   */
  style: _react2.default.PropTypes.object
};

var defaultProps = {
  inset: false
};

var Divider = function Divider(_ref) {
  var inset = _ref.inset;
  var muiTheme = _ref.muiTheme;
  var style = _ref.style;

  var other = _objectWithoutProperties(_ref, ['inset', 'muiTheme', 'style']);

  var styles = {
    root: {
      margin: 0,
      marginTop: -1,
      marginLeft: inset ? 72 : 0,
      height: 1,
      border: 'none',
      backgroundColor: muiTheme.rawTheme.palette.borderColor
    }
  };

  return _react2.default.createElement('hr', _extends({}, other, { style: _styles2.default.prepareStyles(muiTheme, styles.root, style) }));
};

Divider.displayName = 'Divider';
Divider.propTypes = propTypes;
Divider.defaultProps = defaultProps;
Divider = (0, _muiThemeable2.default)(Divider);

exports.default = Divider;
module.exports = exports['default'];