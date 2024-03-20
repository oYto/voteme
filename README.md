3.19：大致完成功能，获取票据和查询票数在没有缓存的情况下，达到800qps；投票接口使用redis分布式锁实现，测试下来只有200qps，打算明天试一试其他方案。

3.20：测试了 mysql 自带的锁和 我们自己实现的乐观锁，发现 qps 没有明显上升，并且由于这两种情况会导致 mysql 压力大，会出现大量慢sql，目前选择继续使用redis。

项目结构整理，查询获票接口大概 400qps，获取票据接口大概 800pqs，投票接口大概200pqs；下阶段考虑先优化投票接口，尝试引入MQ试试。

着手引入MQ了，代码写到一半，发现引入MQ的代价太大了，并且项目的复杂度会变高很多，会需要考虑更多问题，并且最终还是需要将数据持久化到数据库中，时间不是很充分，决定还是在原有基础上进行优化吧。

目前能想到的是，通过 redis 定期缓存所有选手的投票情况，然后采用 “缓存过期” 策略进行缓存的更新，即缓存失效后，再从数据库拿新数据进行更新。因为对于查询目前投票数这种情况，对于数据的实时性要求不高，只需要做到最终一致性即可。

写redis缓存过程中，发现投票的次数超过上限任然可以使用，这里解决了一下bug，将两个sql合并成一个，并且意外提升了性能。投票性能达到 400+qps。
计划：（1）明天通过引入redis缓存所有选手的获得票数，来提升查询票数的qps；
（2）再通过将扣减使用次数这个操作，放到redis中去做，减少我们的mysql数据库压力，进一步提升pqs。