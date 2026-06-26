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

import React, { useEffect, useState } from 'react';
import {
  API,
  getLogo,
  showError,
  showInfo,
  showSuccess,
  getSystemName,
} from '../../helpers';
import Turnstile from 'react-turnstile';
import { Button, Card, Form, Typography } from '@douyinfe/semi-ui';
import { IconComment, IconLock, IconPhone } from '@douyinfe/semi-icons';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const { Text, Title } = Typography;

const PasswordResetForm = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [inputs, setInputs] = useState({
    phone_number: '',
    sms_code: '',
    password: '',
    password2: '',
  });
  const { phone_number, sms_code, password, password2 } = inputs;

  const [loading, setLoading] = useState(false);
  const [smsCodeLoading, setSmsCodeLoading] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(60);

  const logo = getLogo();
  const systemName = getSystemName();

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
  }, []);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown((countdown) => countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(60);
    }
    return () => clearInterval(countdownInterval);
  }, [disableButton, countdown]);

  function handleChange(name, value) {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  }

  const sendPhoneVerificationCode = async () => {
    if (!phone_number) {
      showError(t('请输入手机号'));
      return;
    }
    if (turnstileEnabled && turnstileToken === '') {
      showInfo(t('请稍后几秒重试，Turnstile 正在检查用户环境！'));
      return;
    }
    setSmsCodeLoading(true);
    try {
      const res = await API.post(
        `/api/user/phone/verification?turnstile=${turnstileToken}`,
        {
          phone_number,
          purpose: 'sms_password_reset',
        },
      );
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('短信验证码已发送'));
        setDisableButton(true);
        setCountdown(60);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('发送短信验证码失败，请重试'));
    } finally {
      setSmsCodeLoading(false);
    }
  };

  async function handleSubmit(e) {
    e?.preventDefault?.();
    if (!phone_number) {
      showError(t('请输入手机号'));
      return;
    }
    if (!sms_code) {
      showError(t('请输入短信验证码'));
      return;
    }
    if (!password) {
      showError(t('请输入新密码'));
      return;
    }
    if (password.length < 8 || password.length > 20) {
      showError(t('密码长度必须为 8 到 20 位'));
      return;
    }
    if (password !== password2) {
      showError(t('两次输入的密码不一致'));
      return;
    }
    if (turnstileEnabled && turnstileToken === '') {
      showInfo(t('请稍后几秒重试，Turnstile 正在检查用户环境！'));
      return;
    }
    setLoading(true);
    try {
      const res = await API.post('/api/user/reset', {
        phone_number,
        sms_code,
        password,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('密码重置成功，请使用新密码登录'));
        navigate('/login');
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('重置密码失败，请重试'));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className='classic-page-fill relative overflow-hidden bg-gray-100 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8'>
      {/* 背景模糊晕染球 */}
      <div
        className='blur-ball blur-ball-indigo'
        style={{ top: '-80px', right: '-80px', transform: 'none' }}
      />
      <div
        className='blur-ball blur-ball-teal'
        style={{ top: '50%', left: '-120px' }}
      />
      <div className='w-full max-w-sm mt-[60px]'>
        <div className='flex flex-col items-center'>
          <div className='w-full max-w-md'>
            <div className='flex items-center justify-center mb-6 gap-2'>
              <img src={logo} alt='Logo' className='h-10 rounded-full' />
              <Title heading={3} className='!text-gray-800'>
                {systemName}
              </Title>
            </div>

            <Card className='border-0 !rounded-2xl overflow-hidden'>
              <div className='flex justify-center pt-6 pb-2'>
                <Title heading={3} className='text-gray-800 dark:text-gray-200'>
                  {t('密码重置')}
                </Title>
              </div>
              <div className='px-2 py-8'>
                <Form className='space-y-3' onSubmit={handleSubmit}>
                  <Form.Input
                    field='phone_number'
                    label={t('手机号')}
                    placeholder={t('请输入手机号')}
                    name='phone_number'
                    value={phone_number}
                    onChange={(value) => handleChange('phone_number', value)}
                    prefix={<IconPhone />}
                  />

                  <Form.Input
                    field='sms_code'
                    label={t('短信验证码')}
                    placeholder={t('请输入短信验证码')}
                    name='sms_code'
                    value={sms_code}
                    onChange={(value) => handleChange('sms_code', value)}
                    prefix={<IconComment />}
                    suffix={
                      <Button
                        onClick={sendPhoneVerificationCode}
                        loading={smsCodeLoading}
                        disabled={disableButton || smsCodeLoading}
                      >
                        {disableButton
                          ? `${t('重新发送')} (${countdown})`
                          : t('获取验证码')}
                      </Button>
                    }
                  />

                  <Form.Input
                    field='password'
                    label={t('新密码')}
                    placeholder={t('输入密码，最短 8 位，最长 20 位')}
                    name='password'
                    value={password}
                    mode='password'
                    onChange={(value) => handleChange('password', value)}
                    prefix={<IconLock />}
                  />

                  <Form.Input
                    field='password2'
                    label={t('确认新密码')}
                    placeholder={t('请再次输入新密码')}
                    name='password2'
                    value={password2}
                    mode='password'
                    onChange={(value) => handleChange('password2', value)}
                    prefix={<IconLock />}
                  />

                  <div className='space-y-2 pt-2'>
                    <Button
                      theme='solid'
                      className='w-full !rounded-full'
                      type='primary'
                      htmlType='submit'
                      loading={loading}
                    >
                      {t('重置密码')}
                    </Button>
                  </div>
                </Form>

                <div className='mt-6 text-center text-sm'>
                  <Text>
                    {t('想起来了？')}{' '}
                    <Link
                      to='/login'
                      className='text-blue-600 hover:text-blue-800 font-medium'
                    >
                      {t('登录')}
                    </Link>
                  </Text>
                </div>
              </div>
            </Card>

            {turnstileEnabled && (
              <div className='flex justify-center mt-6'>
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={(token) => {
                    setTurnstileToken(token);
                  }}
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default PasswordResetForm;
