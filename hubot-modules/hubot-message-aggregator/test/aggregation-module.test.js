const rewire = require('rewire');
const {
    expect
} = require('chai');
const sinon = require('sinon');
const aggregator = rewire('../src/aggregator.js');
const {
    handleReaction,
    handleReactionWithChannelId,
    cleanupBrain
} = aggregator;

describe('reaction-aggregator module exports', () => {
    let robot;
    let res;
    let clock;

    beforeEach(() => {
        // Minimal robot stub
        robot = {
            logger: {
                error: sinon.spy(),
                info: sinon.spy()
            },
            brain: {
                data: {},
                get(key) {
                    return this.data[key];
                },
                set(key, value) {
                    this.data[key] = value;
                },
                remove(key) {
                    delete this.data[key];
                }
            },
            messageRoom: sinon.spy(),
        };
        // Default response stub
        res = {
            message: {},
            send: sinon.spy()
        };
        // Freeze time for deterministic timing-based logic
        clock = sinon.useFakeTimers(new Date('2025-04-24T12:00:00Z').getTime());
    });

    afterEach(() => {
        sinon.restore();
        clock.restore();
        delete process.env.HUBOT_AGGREGATION_CHANNEL;
        delete process.env.HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS;
    });

    describe('handleReaction', () => {
        it('logs an error when aggregation channel is unset', async () => {
            delete process.env.HUBOT_AGGREGATION_CHANNEL;
            await handleReaction(res, robot);
            expect(robot.logger.error.calledOnce).to.be.true;
            expect(robot.logger.error.firstCall.args[0]).to.match(/HUBOT_AGGREGATION_CHANNEL/);
        });

        it('uses findChannelIdByName when channel name provided', async () => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'general';
            const findStub = sinon.stub().resolves('C123');
            aggregator.__set__('findChannelIdByName', findStub);
            await handleReaction(res, robot);
            expect(findStub.calledWith(robot, 'general')).to.be.true;
        });
    });

    describe('handleReactionWithChannelId', () => {
        const baseRes = {
            message: {
                reaction: 'thank',
                type: 'added',
                item: {
                    channel: 'CABC',
                    ts: '123.ts'
                }
            },
            send: sinon.spy()
        };

        beforeEach(() => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'CDEST';
        });

        it('ignores non-matching reactions', () => {
            const wrong = {
                message: {
                    reaction: 'thumbs_up',
                    type: 'added',
                    item: {}
                }
            };
            handleReactionWithChannelId(wrong, robot, 'CDEST', 'thank', 'i');
            expect(robot.messageRoom.notCalled).to.be.true;
        });

        it('skips if message is in aggregation channel', () => {
            const sameChannel = {
                message: {
                    reaction: 'thank',
                    type: 'added',
                    item: {
                        channel: 'CDEST',
                        ts: '1'
                    }
                }
            };
            handleReactionWithChannelId(sameChannel, robot, 'CDEST', 'thank', 'i');
            expect(robot.logger.info.calledWithMatch(/Skipping posting permalink/)).to.be.true;
        });

        it('posts permalink and updates brain when first seen', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            // allow promise resolution
            await Promise.resolve();
            expect(fetchStub.calledWith(robot, 'CABC', '123.ts')).to.be.true;
            expect(robot.messageRoom.calledWith('CDEST', 'http://perma')).to.be.true;
        });

        it('does not repost within 24h', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            const key = 'permalink_http://perma';
            robot.brain.set(key, Date.now());
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.notCalled).to.be.true;
        });

        it('skips private when configured not to', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: true
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            process.env.HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS = 'false';
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.notCalled).to.be.true;
        });

        it('allows private when configured to', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: true
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            process.env.HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS = 'true';
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.calledWith('CDEST', 'http://perma')).to.be.true;
        });
    });

    describe('cleanupBrain', () => {
        it('removes entries older than 24h and retains recent ones', () => {
            const now = Date.now();
            robot.brain.data = {
                'permalink_old': now - 25 * 3600 * 1000,
                'permalink_new': now - 1 * 3600 * 1000
            };
            cleanupBrain(robot);
            expect(robot.brain.data).to.not.have.property('permalink_old');
            expect(robot.brain.data).to.have.property('permalink_new');
        });
    });
});