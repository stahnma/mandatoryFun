const rewire = require('rewire');
const {
    expect
} = require('chai');
const sinon = require('sinon');
const aggregator = rewire('../src/aggregator.js');
const {
    handleReaction,
    handleReactionWithChannelId,
    cleanupBrain,
    findChannelIdByName,
    fetchMessagePermalink
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
        // Inject the faked Date and setInterval into the rewired module (rewire
        // gives the module its own bindings that sinon.useFakeTimers does not replace)
        aggregator.__set__('Date', Date);
        aggregator.__set__('setInterval', setInterval);
    });

    afterEach(() => {
        sinon.restore();
        clock.restore();
        delete process.env.HUBOT_AGGREGATION_CHANNEL;
        delete process.env.HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS;
        delete process.env.HUBOT_AGGREGATION_PATTERN;
        delete process.env.HUBOT_SLACK_TOKEN;
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

        it('calls handleReactionWithChannelId directly when channel ID format provided', async () => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'CABCDEFGH1';
            const handleStub = sinon.stub();
            aggregator.__set__('handleReactionWithChannelId', handleStub);
            await handleReaction(res, robot);
            expect(handleStub.calledOnce).to.be.true;
            expect(handleStub.firstCall.args[2]).to.equal('CABCDEFGH1');
        });

        it('logs error when findChannelIdByName returns null', async () => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'nonexistent';
            const findStub = sinon.stub().resolves(null);
            aggregator.__set__('findChannelIdByName', findStub);
            await handleReaction(res, robot);
            // Allow promise chain to resolve
            await Promise.resolve();
            expect(robot.logger.error.calledOnce).to.be.true;
            expect(robot.logger.error.firstCall.args[0]).to.include('nonexistent');
        });

        it('logs error when findChannelIdByName rejects', async () => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'broken';
            const findStub = sinon.stub().rejects(new Error('API failure'));
            aggregator.__set__('findChannelIdByName', findStub);
            await handleReaction(res, robot);
            // Allow promise chain to resolve
            await Promise.resolve();
            expect(robot.logger.error.calledOnce).to.be.true;
            expect(robot.logger.error.firstCall.args[0]).to.include('Error finding channel ID');
        });

        it('passes custom HUBOT_AGGREGATION_PATTERN to handler', async () => {
            process.env.HUBOT_AGGREGATION_CHANNEL = 'CABCDEFGH1';
            process.env.HUBOT_AGGREGATION_PATTERN = 'kudos';
            const handleStub = sinon.stub();
            aggregator.__set__('handleReactionWithChannelId', handleStub);
            await handleReaction(res, robot);
            expect(handleStub.firstCall.args[3]).to.equal('kudos');
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

        it('ignores reaction type removed', () => {
            const removed = {
                message: {
                    reaction: 'thank',
                    type: 'removed',
                    item: {
                        channel: 'CABC',
                        ts: '123.ts'
                    }
                }
            };
            handleReactionWithChannelId(removed, robot, 'CDEST', 'thank', 'i');
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

        it('stores permalink timestamp in brain after posting', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.brain.data['permalink_http://perma']).to.equal(Date.now());
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

        it('reposts after 24h have elapsed', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            const key = 'permalink_http://perma';
            // Set timestamp to exactly 24 hours ago
            robot.brain.set(key, Date.now() - 24 * 60 * 60 * 1000);
            await handleReactionWithChannelId(baseRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.calledWith('CDEST', 'http://perma')).to.be.true;
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

        it('skips private by default when env var is not set', async () => {
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: true
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            delete process.env.HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS;
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

        it('matches reactions case-insensitively', async () => {
            const upperRes = {
                message: {
                    reaction: 'THANKYOU',
                    type: 'added',
                    item: { channel: 'CABC', ts: '123.ts' }
                },
                send: sinon.spy()
            };
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            await handleReactionWithChannelId(upperRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.calledOnce).to.be.true;
        });

        it('matches partial reaction names via regex', async () => {
            const partialRes = {
                message: {
                    reaction: 'thankful_heart',
                    type: 'added',
                    item: { channel: 'CABC', ts: '123.ts' }
                },
                send: sinon.spy()
            };
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            await handleReactionWithChannelId(partialRes, robot, 'CDEST', 'thank', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.calledOnce).to.be.true;
        });

        it('works with a custom aggregation pattern', async () => {
            const kudosRes = {
                message: {
                    reaction: 'kudos',
                    type: 'added',
                    item: { channel: 'CABC', ts: '123.ts' }
                },
                send: sinon.spy()
            };
            const fetchStub = sinon.stub().resolves({
                permalink: 'http://perma',
                isPrivate: false
            });
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            await handleReactionWithChannelId(kudosRes, robot, 'CDEST', 'kudos', 'i');
            await Promise.resolve();
            expect(robot.messageRoom.calledOnce).to.be.true;
        });

        it('sends error message when fetchMessagePermalink rejects', async () => {
            const fetchStub = sinon.stub().rejects(new Error('Slack API down'));
            aggregator.__set__('fetchMessagePermalink', fetchStub);
            const errorRes = {
                message: {
                    reaction: 'thank',
                    type: 'added',
                    item: { channel: 'CABC', ts: '123.ts' }
                },
                send: sinon.spy()
            };
            handleReactionWithChannelId(errorRes, robot, 'CDEST', 'thank', 'i');
            // Flush the microtask queue so the .catch() handler runs
            await Promise.resolve();
            await Promise.resolve();
            await Promise.resolve();
            expect(robot.logger.error.calledOnce).to.be.true;
            expect(errorRes.send.calledOnce).to.be.true;
            expect(errorRes.send.firstCall.args[0]).to.include('error');
        });
    });

    describe('fetchMessagePermalink', () => {
        let WebClientStub;
        let conversationsInfoStub;
        let chatGetPermalinkStub;

        beforeEach(() => {
            process.env.HUBOT_SLACK_TOKEN = 'xoxb-test-token';
            conversationsInfoStub = sinon.stub();
            chatGetPermalinkStub = sinon.stub();
            WebClientStub = sinon.stub().returns({
                conversations: { info: conversationsInfoStub },
                chat: { getPermalink: chatGetPermalinkStub }
            });
            aggregator.__set__('require', function(mod) {
                if (mod === '@slack/web-api') {
                    return { WebClient: WebClientStub };
                }
                return require(mod);
            });
        });

        it('resolves with permalink and isPrivate for public channel', async () => {
            conversationsInfoStub.resolves({
                ok: true,
                channel: { is_private: false }
            });
            chatGetPermalinkStub.resolves({
                ok: true,
                permalink: 'https://slack.com/archives/C123/p123'
            });
            const result = await fetchMessagePermalink(robot, 'C123', '123.456');
            expect(result.permalink).to.equal('https://slack.com/archives/C123/p123');
            expect(result.isPrivate).to.be.false;
        });

        it('resolves with isPrivate true for private channel', async () => {
            conversationsInfoStub.resolves({
                ok: true,
                channel: { is_private: true }
            });
            chatGetPermalinkStub.resolves({
                ok: true,
                permalink: 'https://slack.com/archives/G123/p123'
            });
            const result = await fetchMessagePermalink(robot, 'G123', '123.456');
            expect(result.isPrivate).to.be.true;
        });

        it('rejects when conversations.info returns ok false', async () => {
            conversationsInfoStub.resolves({ ok: false });
            try {
                await fetchMessagePermalink(robot, 'C123', '123.456');
                expect.fail('should have rejected');
            } catch (err) {
                expect(err.message).to.include('conversation information');
            }
        });

        it('rejects when chat.getPermalink returns ok false', async () => {
            conversationsInfoStub.resolves({
                ok: true,
                channel: { is_private: false }
            });
            chatGetPermalinkStub.resolves({ ok: false });
            try {
                await fetchMessagePermalink(robot, 'C123', '123.456');
                expect.fail('should have rejected');
            } catch (err) {
                expect(err.message).to.include('permalink');
            }
        });

        it('rejects when conversations.info throws', async () => {
            conversationsInfoStub.rejects(new Error('network error'));
            try {
                await fetchMessagePermalink(robot, 'C123', '123.456');
                expect.fail('should have rejected');
            } catch (err) {
                expect(err.message).to.equal('network error');
            }
        });

        it('rejects when chat.getPermalink throws', async () => {
            conversationsInfoStub.resolves({
                ok: true,
                channel: { is_private: false }
            });
            chatGetPermalinkStub.rejects(new Error('permalink API error'));
            try {
                await fetchMessagePermalink(robot, 'C123', '123.456');
                expect.fail('should have rejected');
            } catch (err) {
                expect(err.message).to.equal('permalink API error');
            }
        });

        it('passes correct channel and ts to Slack APIs', async () => {
            conversationsInfoStub.resolves({
                ok: true,
                channel: { is_private: false }
            });
            chatGetPermalinkStub.resolves({
                ok: true,
                permalink: 'https://slack.com/archives/CFOO/p999'
            });
            await fetchMessagePermalink(robot, 'CFOO', '999.111');
            expect(conversationsInfoStub.calledWith({ channel: 'CFOO' })).to.be.true;
            expect(chatGetPermalinkStub.calledWith({ channel: 'CFOO', message_ts: '999.111' })).to.be.true;
        });
    });

    describe('findChannelIdByName', () => {
        let WebClientStub;
        let conversationsListStub;

        beforeEach(() => {
            process.env.HUBOT_SLACK_TOKEN = 'xoxb-test-token';
            conversationsListStub = sinon.stub();
            WebClientStub = sinon.stub().returns({
                conversations: { list: conversationsListStub }
            });
            aggregator.__set__('require', function(mod) {
                if (mod === '@slack/web-api') {
                    return { WebClient: WebClientStub };
                }
                return require(mod);
            });
        });

        it('returns channel ID for matching non-archived channel', async () => {
            conversationsListStub.resolves({
                ok: true,
                channels: [
                    { id: 'C111', name: 'general', is_archived: false },
                    { id: 'C222', name: 'random', is_archived: false }
                ]
            });
            const id = await findChannelIdByName(robot, 'general');
            expect(id).to.equal('C111');
        });

        it('skips archived channels', async () => {
            conversationsListStub.resolves({
                ok: true,
                channels: [
                    { id: 'C111', name: 'old-channel', is_archived: true },
                    { id: 'C222', name: 'old-channel', is_archived: false }
                ]
            });
            const id = await findChannelIdByName(robot, 'old-channel');
            expect(id).to.equal('C222');
        });

        it('returns null when no channel matches', async () => {
            conversationsListStub.resolves({
                ok: true,
                channels: [
                    { id: 'C111', name: 'general', is_archived: false }
                ]
            });
            const id = await findChannelIdByName(robot, 'nonexistent');
            expect(id).to.be.null;
        });

        it('returns null when all matching channels are archived', async () => {
            conversationsListStub.resolves({
                ok: true,
                channels: [
                    { id: 'C111', name: 'target', is_archived: true }
                ]
            });
            const id = await findChannelIdByName(robot, 'target');
            expect(id).to.be.null;
        });

        it('returns null and logs error when conversations.list returns ok false', async () => {
            conversationsListStub.resolves({ ok: false });
            const id = await findChannelIdByName(robot, 'general');
            expect(id).to.be.null;
            expect(robot.logger.error.calledOnce).to.be.true;
        });

        it('returns null and logs error on API failure', async () => {
            conversationsListStub.rejects(new Error('rate limited'));
            const id = await findChannelIdByName(robot, 'general');
            expect(id).to.be.null;
            expect(robot.logger.error.calledOnce).to.be.true;
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

        it('removes entries exactly 24h old', () => {
            const now = Date.now();
            robot.brain.data = {
                'permalink_boundary': now - 24 * 3600 * 1000
            };
            cleanupBrain(robot);
            expect(robot.brain.data).to.not.have.property('permalink_boundary');
        });

        it('preserves non-permalink brain keys', () => {
            const now = Date.now();
            robot.brain.data = {
                'permalink_old': now - 25 * 3600 * 1000,
                'some_other_key': 'important data',
                'user_prefs': { theme: 'dark' }
            };
            cleanupBrain(robot);
            expect(robot.brain.data).to.not.have.property('permalink_old');
            expect(robot.brain.data).to.have.property('some_other_key');
            expect(robot.brain.data).to.have.property('user_prefs');
        });

        it('handles empty brain data', () => {
            robot.brain.data = {};
            cleanupBrain(robot);
            expect(Object.keys(robot.brain.data)).to.have.length(0);
        });
    });

    describe('module initialization', () => {
        it('does not register listener for non-slack adapter', () => {
            const initRobot = {
                adapterName: 'shell',
                hearReaction: sinon.spy()
            };
            aggregator(initRobot);
            expect(initRobot.hearReaction.notCalled).to.be.true;
        });

        it('registers hearReaction for slack adapter', () => {
            const initRobot = {
                adapterName: 'slack',
                hearReaction: sinon.spy(),
                logger: { error: sinon.spy(), info: sinon.spy() },
                brain: { data: {} }
            };
            aggregator(initRobot);
            expect(initRobot.hearReaction.calledOnce).to.be.true;
        });

        it('runs brain cleanup on 24-hour interval', () => {
            const initRobot = {
                adapterName: 'slack',
                hearReaction: sinon.spy(),
                logger: { error: sinon.spy(), info: sinon.spy() },
                brain: {
                    data: {},
                    get(key) { return this.data[key]; },
                    set(key, value) { this.data[key] = value; },
                    remove(key) { delete this.data[key]; }
                }
            };
            aggregator(initRobot);
            // After ticking 24h, Date.now() will be T+24h.
            // "stale" should be older than 24h relative to T+24h → set to T (age = 24h, removed)
            // "fresh" should be younger than 24h relative to T+24h → set to T+12h (age = 12h, kept)
            const now = Date.now();
            initRobot.brain.data['permalink_stale'] = now;
            initRobot.brain.data['permalink_fresh'] = now + 12 * 3600 * 1000;
            // Advance clock by 24 hours to trigger the interval
            clock.tick(24 * 60 * 60 * 1000);
            expect(initRobot.brain.data).to.not.have.property('permalink_stale');
            expect(initRobot.brain.data).to.have.property('permalink_fresh');
        });
    });
});