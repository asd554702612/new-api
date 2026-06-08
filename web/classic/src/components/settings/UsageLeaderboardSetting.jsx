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
  Card,
  Empty,
  Form,
  Input,
  List,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';

const { Text } = Typography;

const UsageLeaderboardSetting = ({ value, onSave, saving }) => {
  const { t } = useTranslation();
  const initialIds = useMemo(() => parseIgnoredUserIds(value), [value]);
  const [selectedIds, setSelectedIds] = useState(initialIds);
  const [keyword, setKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [results, setResults] = useState([]);
  const [knownUsers, setKnownUsers] = useState({});

  useEffect(() => {
    setSelectedIds(initialIds);
  }, [initialIds]);

  const serialized = serializeIgnoredUserIds(selectedIds);
  const hasChanges = serialized !== serializeIgnoredUserIds(initialIds);

  const searchUsers = async () => {
    const text = keyword.trim();
    if (!text) {
      setResults([]);
      return;
    }
    setSearching(true);
    try {
      const res = await API.get('/api/user/search', {
        params: {
          keyword: text,
          group: '',
          p: 1,
          page_size: 8,
        },
      });
      const { success, message, data } = res.data;
      if (!success) {
        showError(message || t('搜索用户失败'));
        return;
      }
      const users = data?.items || data || [];
      setResults(Array.isArray(users) ? users : []);
      setKnownUsers((prev) => {
        const next = { ...prev };
        if (Array.isArray(users)) {
          users.forEach((user) => {
            next[user.id] = user;
          });
        }
        return next;
      });
    } catch (error) {
      showError(error);
    } finally {
      setSearching(false);
    }
  };

  const addUser = (user) => {
    setKnownUsers((prev) => ({
      ...prev,
      [user.id]: user,
    }));
    setSelectedIds((prev) => {
      if (prev.includes(user.id)) return prev;
      return [...prev, user.id];
    });
  };

  const removeUser = (id) => {
    setSelectedIds((prev) => prev.filter((item) => item !== id));
  };

  const save = async () => {
    await onSave(serialized);
  };

  return (
    <Card>
      <Form.Section
        text={t('排行榜设置')}
        extraText={t('配置用户消耗榜单中需要忽略的用户')}
      >
        <Space vertical align='stretch' style={{ width: '100%' }}>
          <Space>
            <Input
              value={keyword}
              placeholder={t('按名称、邮箱或 ID 搜索用户')}
              onChange={setKeyword}
              onEnterPress={searchUsers}
              style={{ width: 320 }}
            />
            <Button
              icon={<IconSearch />}
              loading={searching}
              onClick={searchUsers}
            >
              {t('搜索')}
            </Button>
          </Space>

          {results.length > 0 && (
            <List
              bordered
              dataSource={results}
              renderItem={(user) => {
                const selected = selectedIds.includes(user.id);
                return (
                  <List.Item
                    main={
                      <div>
                        <Text strong>{userLabel(user)}</Text>
                        <div>
                          <Text type='tertiary' size='small'>
                            ID: {user.id}
                          </Text>
                        </div>
                      </div>
                    }
                    extra={
                      <Button
                        size='small'
                        disabled={selected}
                        onClick={() => addUser(user)}
                      >
                        {selected ? t('已选择') : t('添加')}
                      </Button>
                    }
                  />
                );
              }}
            />
          )}

          <div>
            <Text strong>{t('忽略用户')}</Text>
            <div style={{ marginTop: 8 }}>
              {selectedIds.length === 0 ? (
                <Empty title={t('暂无忽略用户')} />
              ) : (
                <Space wrap>
                  {selectedIds.map((id) => (
                    <Tag
                      key={id}
                      closable
                      onClose={() => removeUser(id)}
                      color='blue'
                    >
                      {knownUsers[id]
                        ? userLabel(knownUsers[id])
                        : `User #${id}`}
                    </Tag>
                  ))}
                </Space>
              )}
            </div>
          </div>

          <Space>
            <Button
              type='primary'
              loading={saving}
              disabled={!hasChanges}
              onClick={save}
            >
              {t('保存排行榜设置')}
            </Button>
            <Button
              disabled={!hasChanges || saving}
              onClick={() => setSelectedIds(initialIds)}
            >
              {t('重置')}
            </Button>
          </Space>
        </Space>
      </Form.Section>
    </Card>
  );
};

function parseIgnoredUserIds(raw) {
  if (!raw || !String(raw).trim()) return [];
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    const seen = new Set();
    return parsed
      .map((item) => Number(item))
      .filter((id) => {
        if (!Number.isInteger(id) || id <= 0 || seen.has(id)) return false;
        seen.add(id);
        return true;
      });
  } catch (error) {
    return [];
  }
}

function serializeIgnoredUserIds(ids) {
  return JSON.stringify(parseIgnoredUserIds(JSON.stringify(ids)));
}

function userLabel(user) {
  const name = user.display_name || user.username || `User #${user.id}`;
  if (user.email) return `${name} (${user.email})`;
  return name;
}

export default UsageLeaderboardSetting;
