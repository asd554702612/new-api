/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

export const PROJECT_REPOSITORY_URL = 'https://github.com/QuantumNous/new-api';
export const DEFAULT_DOC_URL = 'https://docs.newapi.pro';

export const getHomeLandingData = ({ t, docsLink, systemName }) => {
  const docUrl = docsLink || DEFAULT_DOC_URL;

  return {
    docUrl,
    logoTagline: t('企业级国际 API 中转服务'),
    navItems: [
      { href: '#top', label: t('首页') },
      { href: '#pricing', label: t('定价') },
      { href: docUrl, label: t('API 文档'), external: true },
      { href: '/feedback', label: t('投诉反馈') },
      { href: '#footer', label: t('技术社群') },
    ],
    hero: {
      badge: t('稳定 · 安全 · 高效的国际中转服务'),
      titlePrefix: systemName,
      titleLead: t('让 '),
      titleHighlight: t('AI 开发'),
      titleSuffix: t('更简单'),
      subtitle: t('企业级安全接入'),
      primaryGuest: t('立即开始接入'),
      primaryAuthed: t('立即使用'),
      secondary: t('查看文档'),
      checks: [t('全链路加速'), t('弹性调度'), t('TLS 加密')],
      stats: [
        { value: '1.2M+', label: t('API 调用次数') },
        { value: '50,000+', label: t('开发者信赖') },
        { value: '99.9%', label: t('可用性保障') },
        { value: '24/7', label: t('技术支持') },
      ],
    },
    features: {
      kicker: `${t('为什么选择')} ${systemName}`,
      title: t('快速、稳定、安全的 AI 接入体验'),
      subtitle: t(
        '我们提供稳定、快速、安全的国际 API 中转服务，让业务专注于产品创新。',
      ),
      cards: [
        {
          icon: 'shield',
          title: t('企业级安全'),
          description: t('全程 TLS 加密与独立通道设计，兼顾安全与合规。'),
        },
        {
          icon: 'swap',
          title: t('简单易用'),
          description: t('兼容国际接口规范，替换 Base URL 即可接入。'),
        },
        {
          icon: 'chartBar',
          title: t('高速稳定'),
          description: t('全链路加速与弹性调度，保障高峰期也能稳定返回。'),
        },
      ],
    },
    products: {
      kicker: t('核心能力矩阵'),
      title: t('一套入口，承接你的 AI 业务增长'),
      subtitle: t(
        '从接入、路由到交付与安全，围绕真实生产环境设计，适合个人开发者、创业团队与企业项目逐步扩展。',
      ),
      sideTitle: t('覆盖当前 API 可用的主流推理、轻量与图像场景。'),
      sideSubtitle: t(
        '按场景选择高性能、均衡、低成本或图像能力，统一通过 {{systemName}} 接入。',
        { systemName },
      ),
      cards: [
        {
          icon: 'cpu',
          title: t('统一模型接入'),
          description: t(
            '集中承接多模型能力，让团队用一致方式访问推理、图像与扩展能力。',
          ),
        },
        {
          icon: 'sparkles',
          title: t('低成本迁移上线'),
          description: t(
            '沿用熟悉的国际请求方式，把改造成本降到更低，快速完成联调与交付。',
          ),
        },
      ],
      models: [
        'DeepSeek V4-Pro',
        'DeepSeek V4-Flash',
        'Doubao-Seed_1.5',
        'Doubao-Seed_1.6',
        'Doubao-Seed_1.8',
        'Doubao-Seed_2.0',
        'Qwen3.5-Plus',
        'Qwen',
        'Happy horse',
        'Kling Video API',
        'MiniMax-M2.7',
        'MiniMax-M2.5',
        'GLM-5.1',
        'GLM-5-Turbo',
        'GLM-5',
      ],
    },
    pricing: {
      kicker: t('灵活的计费方案'),
      title: t('灵活套餐，按需选择'),
      subtitle: t('按量计费、轻量订阅与企业级方案都能覆盖。'),
      viewDetails: t('查看详细价格'),
      rows: [
        {
          plan: t('标准套餐'),
          price: t('¥29.90/月'),
          description: t('适合个人开发者的轻度项目'),
          badge: t('轻量版'),
          featured: false,
        },
        {
          plan: t('专业套餐'),
          price: t('¥69.90/月'),
          description: t('适合中型团队和稳定项目'),
          badge: t('推荐'),
          featured: false,
        },
        {
          plan: t('企业套餐'),
          price: t('¥159.90/月'),
          description: t('适合大型企业和定制化需求'),
          badge: t('企业版'),
          featured: true,
        },
      ],
    },
    showcase: {
      kicker: t('信任与案例展示'),
      title: t('从接入效率到稳定运行，都更贴近生产需求'),
      subtitle: t(
        '新版官网保留真实业务表达，用更现代的展示方式承接你的平台优势、客户感知与后续案例内容。',
      ),
      stats: [
        { icon: 'users', value: '50,000+', label: t('开发者信赖') },
        { icon: 'shield', value: '99.9%', label: t('服务可用性') },
        { icon: 'sparkles', value: '24/7', label: t('支持响应') },
      ],
      quotes: [
        {
          quote: t(
            '接入方式几乎不用重写，团队把更多时间放在产品功能上，而不是处理兼容和网络问题。',
          ),
          author: t('产品团队反馈'),
          role: t('企业接入场景'),
        },
        {
          quote: t(
            '从测试到上线节奏更顺，接口稳定之后，前后端协作成本明显下降。',
          ),
          author: t('研发视角'),
          role: t('工程实践体验'),
        },
        {
          quote: t(
            '对我们来说，更重要的是稳定与支持响应速度，这也是长期使用的基础。',
          ),
          author: t('运维协作反馈'),
          role: t('生产可用性关注'),
        },
      ],
    },
    cta: {
      badge: t('立即开始，快速接入'),
      title: t('准备好把你的 AI 应用更快推向生产了吗？'),
      subtitle: t(
        '沿用熟悉的调用方式，保留现有开发习惯，用更稳定的链路支撑接入、测试、上线和长期维护。',
      ),
      trustItems: [
        t('支持试用与快速体验'),
        t('无需复杂迁移流程'),
        t('技术支持响应更直接'),
      ],
    },
    footer: {
      about: t('企业级国际 API 中转服务，让 AI 开发更简单。'),
      follow: t('关注我们'),
      socialDots: [t('微博'), t('视'), t('知'), t('公司')],
      columns: [
        {
          title: t('产品'),
          items: [
            { label: t('定价'), href: '#pricing' },
            { label: t('API 文档'), href: docUrl, external: true },
            { label: t('支持的平台'), href: '#overview' },
            { label: t('更新日志'), href: docUrl, external: true },
          ],
        },
        {
          title: t('开发者'),
          items: [
            { label: t('快速开始'), href: docUrl, external: true },
            { label: t('SDK & 工具'), href: docUrl, external: true },
            { label: t('最佳实践'), href: docUrl, external: true },
            { label: t('API 状态'), href: docUrl, external: true },
          ],
        },
        {
          title: t('公司'),
          items: [
            { label: t('关于我们'), href: '#overview' },
            { label: t('联系我们'), href: '#footer' },
            { label: t('投诉反馈'), href: '/feedback' },
            { label: t('服务条款'), href: docUrl, external: true },
            { label: t('隐私政策'), href: docUrl, external: true },
          ],
        },
      ],
    },
  };
};
