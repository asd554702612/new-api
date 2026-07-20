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

import React, { useEffect, useMemo, useState, useContext } from 'react';
import {
  Button,
  Card,
  Input,
  Select,
  Space,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { ClipboardCheck, MessageSquare, Search } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import Turnstile from 'react-turnstile';
import { API, copy, showError, showSuccess } from '../../helpers';
import { UserContext } from '../../context/User';

const INITIAL_FORM = {
  feedback_type: 'complaint',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  title: '',
  content: '',
};

const readLocalUser = () => {
  try {
    const raw = localStorage.getItem('user');
    return raw ? JSON.parse(raw) : undefined;
  } catch {
    return undefined;
  }
};

const buildUserContactFields = (user) => ({
  contact_name: user?.display_name || user?.displayName || user?.username || '',
  contact_email: user?.email || '',
  contact_phone: user?.phone_number || user?.phone || '',
});

const unwrapApiData = (res) => {
  const body = res?.data;
  return body?.data ?? body;
};

const formatDate = (value) => {
  if (!value) return '-';
  const normalized =
    typeof value === 'number' && value < 10000000000 ? value * 1000 : value;
  const parsed = new Date(normalized);
  return Number.isNaN(parsed.getTime())
    ? String(value)
    : parsed.toLocaleString();
};

const getStatusLabel = (t, status) => {
  const labels = {
    pending: t('待处理'),
    processing: t('处理中'),
    resolved: t('已解决'),
    closed: t('已关闭'),
    rejected: t('已驳回'),
  };
  return labels[status] || status || '-';
};

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    processing: 'blue',
    resolved: 'green',
    closed: 'grey',
    rejected: 'red',
  };
  return colors[status] || 'grey';
};

const getFeedbackTypeLabel = (t, type) => {
  const labels = {
    complaint: t('投诉'),
    feedback: t('反馈'),
    other: t('其他'),
  };
  return labels[type] || type || '-';
};

const Feedback = () => {
  const { t } = useTranslation();
  const [userState] = useContext(UserContext);
  const [form, setForm] = useState(INITIAL_FORM);
  const [loading, setLoading] = useState(false);
  const [trackingCode, setTrackingCode] = useState('');
  const [trackingCodeQuery, setTrackingCodeQuery] = useState('');
  const [trackingLoading, setTrackingLoading] = useState(false);
  const [trackedFeedback, setTrackedFeedback] = useState(null);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const currentUser = userState?.user || readLocalUser();

  const typeOptions = useMemo(
    () => [
      { label: t('投诉'), value: 'complaint' },
      { label: t('反馈'), value: 'feedback' },
      { label: t('其他'), value: 'other' },
    ],
    [t],
  );

  const updateField = (field, value) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  useEffect(() => {
    API.get('/api/status', { disableDuplicate: true })
      .then((res) => {
        const status = res?.data?.data || {};
        setTurnstileEnabled(Boolean(status.turnstile_check));
        setTurnstileSiteKey(status.turnstile_site_key || '');
      })
      .catch(() => {
        setTurnstileEnabled(false);
        setTurnstileSiteKey('');
      });
  }, []);

  useEffect(() => {
    if (!currentUser) {
      return;
    }

    const contactFields = buildUserContactFields(currentUser);
    setForm((prev) => ({
      ...prev,
      contact_name: prev.contact_name || contactFields.contact_name,
      contact_email: prev.contact_email || contactFields.contact_email,
      contact_phone: prev.contact_phone || contactFields.contact_phone,
    }));
  }, [currentUser]);

  const validateForm = () => {
    if (!form.contact_name.trim()) {
      showError(t('请填写联系人'));
      return false;
    }
    if (!form.contact_email.trim() && !form.contact_phone.trim()) {
      showError(t('请至少填写邮箱或手机号'));
      return false;
    }
    if (!form.title.trim()) {
      showError(t('请填写标题'));
      return false;
    }
    if (!form.content.trim()) {
      showError(t('请填写内容'));
      return false;
    }
    if (turnstileEnabled && !turnstileToken) {
      showError(t('请稍后几秒重试，Turnstile 正在检查用户环境！'));
      return false;
    }
    return true;
  };

  const submitFeedback = async () => {
    if (!validateForm()) return;

    setLoading(true);
    try {
      const payload = {
        ...form,
        feedback_type: form.feedback_type,
        contact_name: form.contact_name.trim(),
        contact_email: form.contact_email.trim(),
        contact_phone: form.contact_phone.trim(),
        title: form.title.trim(),
        content: form.content.trim(),
      };
      const res = await API.post('/api/feedback', payload, {
        params: turnstileEnabled ? { turnstile: turnstileToken } : undefined,
      });
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('提交失败，请重试'));
        return;
      }

      const data = unwrapApiData(res) || {};
      const code =
        data.tracking_code ||
        data.trackingCode ||
        res.data?.tracking_code ||
        '';
      setTrackingCode(code);
      setForm((prev) => ({
        ...INITIAL_FORM,
        ...buildUserContactFields(currentUser),
        feedback_type: prev.feedback_type,
      }));
      setTurnstileToken('');
      showSuccess(t('提交成功'));
    } catch (error) {
      showError(error?.message || t('提交失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  const copyTrackingCode = async () => {
    if (await copy(trackingCode)) {
      showSuccess(t('已复制追踪码'));
    }
  };

  const queryTrackingCode = async () => {
    const code = trackingCodeQuery.trim();
    if (!code) {
      showError(t('请填写追踪码'));
      return;
    }

    setTrackingLoading(true);
    try {
      const res = await API.get(
        `/api/feedback/track/${encodeURIComponent(code)}`,
        { disableDuplicate: true },
      );
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('未找到投诉反馈记录'));
        setTrackedFeedback(null);
        return;
      }
      setTrackedFeedback(unwrapApiData(res));
    } catch (error) {
      showError(error?.message || t('查询失败，请重试'));
      setTrackedFeedback(null);
    } finally {
      setTrackingLoading(false);
    }
  };

  return (
    <div className='mt-[60px] px-2'>
      <div className='w-full max-w-3xl mx-auto space-y-5'>
        <Card className='!rounded-2xl shadow-sm border-0'>
          <div className='flex items-center gap-3 mb-5'>
            <div className='w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center'>
              <MessageSquare size={20} />
            </div>
            <div>
              <Typography.Title heading={4} style={{ margin: 0 }}>
                {t('公众投诉反馈')}
              </Typography.Title>
            </div>
          </div>

          {trackingCode && (
            <div className='mb-5 rounded-xl border border-green-200 bg-green-50 p-4'>
              <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3'>
                <div>
                  <div className='font-semibold text-green-700'>
                    {t('提交成功')}
                  </div>
                  <div className='text-sm text-green-700 break-all'>
                    {t('追踪码')}：{trackingCode}
                  </div>
                </div>
                <Button
                  type='primary'
                  theme='outline'
                  icon={<ClipboardCheck size={16} />}
                  onClick={copyTrackingCode}
                >
                  {t('复制追踪码')}
                </Button>
              </div>
            </div>
          )}

          <Space vertical align='start' className='w-full'>
            <div className='grid grid-cols-1 md:grid-cols-2 gap-4 w-full'>
              <div>
                <div className='text-sm font-medium mb-2'>{t('类型')}</div>
                <Select
                  value={form.feedback_type}
                  optionList={typeOptions}
                  onChange={(value) => updateField('feedback_type', value)}
                  style={{ width: '100%' }}
                />
              </div>
              <div>
                <div className='text-sm font-medium mb-2'>{t('联系人')}</div>
                <Input
                  value={form.contact_name}
                  onChange={(value) => updateField('contact_name', value)}
                  placeholder={t('请输入联系人')}
                />
              </div>
              <div>
                <div className='text-sm font-medium mb-2'>{t('邮箱')}</div>
                <Input
                  value={form.contact_email}
                  onChange={(value) => updateField('contact_email', value)}
                  placeholder={t('请输入邮箱')}
                />
              </div>
              <div>
                <div className='text-sm font-medium mb-2'>{t('手机号')}</div>
                <Input
                  value={form.contact_phone}
                  onChange={(value) => updateField('contact_phone', value)}
                  placeholder={t('请输入手机号')}
                />
              </div>
            </div>

            <div className='w-full'>
              <div className='text-sm font-medium mb-2'>{t('标题')}</div>
              <Input
                value={form.title}
                onChange={(value) => updateField('title', value)}
                placeholder={t('请简要描述事项')}
              />
            </div>

            <div className='w-full'>
              <div className='text-sm font-medium mb-2'>{t('内容')}</div>
              <TextArea
                value={form.content}
                onChange={(value) => updateField('content', value)}
                autosize={{ minRows: 6, maxRows: 12 }}
                placeholder={t('请填写具体情况、诉求或建议')}
              />
            </div>

            {turnstileEnabled && turnstileSiteKey && (
              <div className='w-full flex justify-center'>
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={setTurnstileToken}
                  onExpire={() => setTurnstileToken('')}
                />
              </div>
            )}

            <div className='w-full flex justify-end pt-2'>
              <Button type='primary' loading={loading} onClick={submitFeedback}>
                {t('提交')}
              </Button>
            </div>
          </Space>
        </Card>

        <Card className='!rounded-2xl shadow-sm border-0'>
          <div className='flex items-center gap-3 mb-5'>
            <div className='w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center'>
              <Search size={20} />
            </div>
            <Typography.Title heading={4} style={{ margin: 0 }}>
              {t('查询投诉反馈')}
            </Typography.Title>
          </div>

          <Space vertical align='start' className='w-full'>
            <div className='flex flex-col sm:flex-row gap-3 w-full'>
              <Input
                value={trackingCodeQuery}
                onChange={setTrackingCodeQuery}
                placeholder={t('请输入追踪码')}
              />
              <Button
                type='primary'
                theme='outline'
                icon={<Search size={16} />}
                loading={trackingLoading}
                onClick={queryTrackingCode}
              >
                {t('查询')}
              </Button>
            </div>

            {trackedFeedback && (
              <div className='w-full rounded-xl border border-gray-100 p-4'>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-3'>
                  <div>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('追踪码')}
                    </div>
                    <div className='text-sm font-medium break-all'>
                      {trackedFeedback.tracking_code || '-'}
                    </div>
                  </div>
                  <div>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('状态')}
                    </div>
                    <Tag color={getStatusColor(trackedFeedback.status)}>
                      {getStatusLabel(t, trackedFeedback.status)}
                    </Tag>
                  </div>
                  <div>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('类型')}
                    </div>
                    <div className='text-sm'>
                      {getFeedbackTypeLabel(t, trackedFeedback.feedback_type)}
                    </div>
                  </div>
                  <div>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('处理时间')}
                    </div>
                    <div className='text-sm'>
                      {formatDate(trackedFeedback.handled_at)}
                    </div>
                  </div>
                  <div className='md:col-span-2'>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('标题')}
                    </div>
                    <div className='text-sm font-medium break-all'>
                      {trackedFeedback.title || '-'}
                    </div>
                  </div>
                  <div className='md:col-span-2'>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('内容')}
                    </div>
                    <div className='text-sm whitespace-pre-wrap break-all'>
                      {trackedFeedback.content || '-'}
                    </div>
                  </div>
                  <div className='md:col-span-2'>
                    <div className='text-xs text-gray-500 mb-1'>
                      {t('处理说明')}
                    </div>
                    <div className='text-sm whitespace-pre-wrap break-all'>
                      {trackedFeedback.admin_note || '-'}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </Space>
        </Card>
      </div>
    </div>
  );
};

export default Feedback;
