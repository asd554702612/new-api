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
  Input,
  Select,
  Space,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { ClipboardCheck, MessageSquare } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import Turnstile from 'react-turnstile';
import { API, copy, showError, showSuccess } from '../../helpers';

const INITIAL_FORM = {
  feedback_type: 'complaint',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  title: '',
  content: '',
};

const unwrapApiData = (res) => {
  const body = res?.data;
  return body?.data ?? body;
};

const Feedback = () => {
  const { t } = useTranslation();
  const [form, setForm] = useState(INITIAL_FORM);
  const [loading, setLoading] = useState(false);
  const [trackingCode, setTrackingCode] = useState('');
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');

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
      setForm(INITIAL_FORM);
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

  return (
    <div className='mt-[60px] px-2'>
      <div className='w-full max-w-3xl mx-auto'>
        <Card className='!rounded-2xl shadow-sm border-0'>
          <div className='flex items-center gap-3 mb-5'>
            <div className='w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center'>
              <MessageSquare size={20} />
            </div>
            <div>
              <Typography.Title heading={4} style={{ margin: 0 }}>
                {t('公众投诉反馈')}
              </Typography.Title>
              <Typography.Text type='secondary'>
                {t('无需登录即可提交投诉、反馈或其他事项')}
              </Typography.Text>
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
      </div>
    </div>
  );
};

export default Feedback;
