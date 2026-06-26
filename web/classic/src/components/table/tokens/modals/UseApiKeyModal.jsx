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

import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Modal,
  Space,
  TabPane,
  Tabs,
  Typography,
} from '@douyinfe/semi-ui';
import { IconCopy, IconTerminal } from '@douyinfe/semi-icons';

const { Text } = Typography;

function getServerAddress() {
  try {
    const raw = localStorage.getItem('status');
    if (raw) {
      const status = JSON.parse(raw);
      if (status.server_address) return status.server_address;
    }
  } catch (_) {}

  return window.location.origin;
}

function normalizeKey(tokenKey) {
  if (!tokenKey) return '';
  return tokenKey.startsWith('sk-') ? tokenKey : `sk-${tokenKey}`;
}

function buildCodexConfig(serverAddress) {
  return `model_provider = "OpenAI"
model = "gpt-5.5"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${serverAddress}/v1"
wire_api = "responses"
requires_openai_auth = true

[features]
goals = true`;
}

function buildCodexAuth(apiKey) {
  return `{
  "OPENAI_API_KEY": "${apiKey}"
}`;
}

const platformItems = [
  {
    key: 'macos-linux',
    label: 'macOS / Linux',
    configPath: '~/.codex/config.toml',
    authPath: '~/.codex/auth.json',
  },
  {
    key: 'windows',
    label: 'Windows',
    configPath: '%USERPROFILE%\\.codex\\config.toml',
    authPath: '%USERPROFILE%\\.codex\\auth.json',
  },
];

const codeBlockStyle = {
  backgroundColor: 'var(--semi-color-bg-1)',
  borderColor: 'var(--semi-color-border)',
};

const codeBlockHeaderStyle = {
  backgroundColor: 'var(--semi-color-fill-0)',
  borderBottomColor: 'var(--semi-color-border)',
};

const codeBlockTitleStyle = {
  color: 'var(--semi-color-text-2)',
};

const codeBlockCopyStyle = {
  color: 'var(--semi-color-text-1)',
};

const codeTextStyle = {
  color: 'var(--semi-color-text-0)',
};

function CodeBlock({ title, code, onCopy, t }) {
  return (
    <div className='overflow-hidden rounded-xl border' style={codeBlockStyle}>
      <div
        className='flex items-center justify-between border-b px-4 py-2 text-sm'
        style={codeBlockHeaderStyle}
      >
        <span style={codeBlockTitleStyle}>{title}</span>
        <Button
          type='tertiary'
          theme='borderless'
          size='small'
          icon={<IconCopy />}
          onClick={() => onCopy(code)}
          style={codeBlockCopyStyle}
        >
          {t('复制')}
        </Button>
      </div>
      <pre className='m-0 overflow-x-auto whitespace-pre-wrap break-words p-4 font-mono text-sm leading-6'>
        <code style={codeTextStyle}>{code}</code>
      </pre>
    </div>
  );
}

function CodexInstructions({ platform, configText, authText, copyText, t }) {
  return (
    <div className='space-y-4'>
      <Text type='warning'>
        {t('请确保以下内容位于 config.toml 文件的开头部分')}
      </Text>
      <CodeBlock
        title={platform.configPath}
        code={configText}
        onCopy={copyText}
        t={t}
      />
      <CodeBlock
        title={platform.authPath}
        code={authText}
        onCopy={copyText}
        t={t}
      />
    </div>
  );
}

const UseApiKeyModal = ({
  visible,
  onCancel,
  tokenKey,
  initialTab = 'codex-app',
  copyText,
  t,
}) => {
  const [activeTab, setActiveTab] = useState(initialTab);
  const [activePlatform, setActivePlatform] = useState(platformItems[0].key);

  const apiKey = normalizeKey(tokenKey);
  const serverAddress = getServerAddress();
  const configText = useMemo(
    () => buildCodexConfig(serverAddress),
    [serverAddress],
  );
  const authText = useMemo(() => buildCodexAuth(apiKey), [apiKey]);

  const currentPlatform =
    platformItems.find((item) => item.key === activePlatform) ||
    platformItems[0];

  useEffect(() => {
    if (!visible) return;
    setActiveTab(initialTab);
    setActivePlatform(platformItems[0].key);
  }, [initialTab, visible]);

  return (
    <Modal
      title={t('使用 API 密钥')}
      visible={visible}
      onCancel={onCancel}
      afterClose={() => {
        setActiveTab(initialTab);
        setActivePlatform(platformItems[0].key);
      }}
      icon={null}
      footer={
        <Button onClick={onCancel} type='tertiary'>
          {t('关闭')}
        </Button>
      }
      width={1120}
      bodyStyle={{ maxHeight: '72vh', overflowY: 'auto' }}
    >
      <div className='space-y-4'>
        <div className='text-base text-[var(--semi-color-text-1)]'>
          {t(
            '复制并运行对应系统脚本，按提示手动输入 API Key，即可为 Codex CLI 与 Codex App 写入配置。',
          )}
        </div>

        <Tabs type='line' activeKey={activeTab} onChange={setActiveTab}>
          <TabPane
            itemKey='codex-cli'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                {t('Codex CLI')}
              </span>
            }
          >
            <Tabs
              type='line'
              activeKey={activePlatform}
              onChange={setActivePlatform}
            >
              {platformItems.map((platform) => (
                <TabPane
                  key={platform.key}
                  itemKey={platform.key}
                  tab={
                    <span className='inline-flex items-center gap-2'>
                      <IconTerminal />
                      {platform.label}
                    </span>
                  }
                >
                  <CodexInstructions
                    platform={platform}
                    configText={configText}
                    authText={authText}
                    copyText={copyText}
                    t={t}
                  />
                </TabPane>
              ))}
            </Tabs>
          </TabPane>

          <TabPane
            itemKey='codex-app'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                {t('Codex App')}
              </span>
            }
          >
            <div className='space-y-4'>
              <div className='rounded-lg border border-[var(--semi-color-border)] bg-[var(--semi-color-fill-0)] p-4'>
                <ol className='m-0 space-y-2 pl-5 text-[var(--semi-color-text-1)]'>
                  <li>
                    {t(
                      '打开 Codex App 设置，找到并打开用户级 config.toml 配置文件。',
                    )}
                  </li>
                  <li>
                    {t(
                      '将下方 config.toml 内容放在文件开头；如果已有同名配置，请先备份后替换。',
                    )}
                  </li>
                  <li>
                    {t(
                      '将下方 auth.json 内容写入用户级 auth.json，然后重启 Codex App。',
                    )}
                  </li>
                </ol>
              </div>
              <CodexInstructions
                platform={currentPlatform}
                configText={configText}
                authText={authText}
                copyText={copyText}
                t={t}
              />
            </div>
          </TabPane>

          <TabPane
            itemKey='one-click'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                {t('一键配置')}
              </span>
            }
          >
            <div className='rounded-lg border border-[var(--semi-color-border)] bg-[var(--semi-color-fill-0)] p-4 text-[var(--semi-color-text-1)]'>
              {t('暂未提供一键脚本，请使用 Codex CLI 或 Codex App 手动配置。')}
            </div>
          </TabPane>

          <TabPane
            itemKey='codex-cli-websocket'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                {t('Codex CLI (WebSocket)')}
              </span>
            }
          >
            <div className='rounded-lg border border-[var(--semi-color-border)] bg-[var(--semi-color-fill-0)] p-4 text-[var(--semi-color-text-1)]'>
              {t('当前令牌使用 OpenAI 兼容配置；WebSocket 配置请联系管理员。')}
            </div>
          </TabPane>

          <TabPane
            itemKey='claude-code'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                Claude Code
              </span>
            }
          >
            <Space vertical align='start' spacing='medium'>
              <Text>
                {t('Claude Code 可使用 OpenAI 兼容地址与当前 API Key。')}
              </Text>
              <CodeBlock
                title='.env'
                code={`OPENAI_BASE_URL=${serverAddress}/v1\nOPENAI_API_KEY=${apiKey}`}
                onCopy={copyText}
                t={t}
              />
            </Space>
          </TabPane>

          <TabPane
            itemKey='opencode'
            tab={
              <span className='inline-flex items-center gap-2'>
                <IconTerminal />
                OpenCode
              </span>
            }
          >
            <Space vertical align='start' spacing='medium'>
              <Text>
                {t('OpenCode 可使用 OpenAI 兼容地址与当前 API Key。')}
              </Text>
              <CodeBlock
                title='.env'
                code={`OPENAI_BASE_URL=${serverAddress}/v1\nOPENAI_API_KEY=${apiKey}`}
                onCopy={copyText}
                t={t}
              />
            </Space>
          </TabPane>
        </Tabs>
      </div>
    </Modal>
  );
};

export default UseApiKeyModal;
