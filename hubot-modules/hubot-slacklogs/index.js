'use strict'

const path = require('path')

module.exports = (robot) => {
  const scriptsPath = path.resolve(__dirname, 'src')
  robot.loadFile(scriptsPath, 'slacklogs.js')
  robot.loadFile(scriptsPath, 'slackreactions.js')
}
