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

import React from 'react';
import {
  Button,
  Input,
  Modal,
  Radio,
  RadioGroup,
  Typography,
} from '@douyinfe/semi-ui';
import { IconComment, IconLock } from '@douyinfe/semi-icons';
import Turnstile from 'react-turnstile';

const ChangePasswordModal = ({
  t,
  showChangePasswordModal,
  setShowChangePasswordModal,
  inputs,
  handleInputChange,
  changePassword,
  turnstileEnabled,
  turnstileSiteKey,
  setTurnstileToken,
  phoneNumber,
  sendPasswordPhoneVerificationCode,
  disableButton,
  loading,
  countdown,
}) => {
  return (
    <Modal
      title={
        <div className='flex items-center'>
          <IconLock className='mr-2 text-orange-500' />
          {t('修改密码')}
        </div>
      }
      visible={showChangePasswordModal}
      onCancel={() => setShowChangePasswordModal(false)}
      onOk={changePassword}
      size={'small'}
      centered={true}
      className='modern-modal'
    >
      <div className='space-y-4 py-4'>
        <RadioGroup
          type='button'
          value={inputs.password_verification_method}
          onChange={(event) =>
            handleInputChange(
              'password_verification_method',
              event.target.value,
            )
          }
        >
          <Radio value='password'>{t('原密码验证')}</Radio>
          <Radio value='phone' disabled={!phoneNumber}>
            {t('手机号验证')}
          </Radio>
        </RadioGroup>

        {inputs.password_verification_method === 'phone' ? (
          <div>
            <Typography.Text strong className='block mb-2'>
              {t('短信验证码')}
            </Typography.Text>
            <div className='flex gap-2'>
              <Input
                name='password_phone_verification_code'
                placeholder={t('请输入短信验证码')}
                value={inputs.password_phone_verification_code}
                onChange={(value) =>
                  handleInputChange('password_phone_verification_code', value)
                }
                size='large'
                className='!rounded-lg'
                prefix={<IconComment />}
              />
              <Button
                onClick={sendPasswordPhoneVerificationCode}
                disabled={disableButton || loading}
                loading={loading}
                type='primary'
                theme='outline'
              >
                {disableButton
                  ? `${t('重新发送')} (${countdown})`
                  : t('获取验证码')}
              </Button>
            </div>
          </div>
        ) : (
          <div>
            <Typography.Text strong className='block mb-2'>
              {t('原密码')}
            </Typography.Text>
            <Input
              name='original_password'
              placeholder={t('请输入原密码')}
              type='password'
              value={inputs.original_password}
              onChange={(value) =>
                handleInputChange('original_password', value)
              }
              size='large'
              className='!rounded-lg'
              prefix={<IconLock />}
            />
          </div>
        )}

        <div>
          <Typography.Text strong className='block mb-2'>
            {t('新密码')}
          </Typography.Text>
          <Input
            name='set_new_password'
            placeholder={t('请输入新密码')}
            type='password'
            value={inputs.set_new_password}
            onChange={(value) => handleInputChange('set_new_password', value)}
            size='large'
            className='!rounded-lg'
            prefix={<IconLock />}
          />
        </div>

        <div>
          <Typography.Text strong className='block mb-2'>
            {t('确认新密码')}
          </Typography.Text>
          <Input
            name='set_new_password_confirmation'
            placeholder={t('请再次输入新密码')}
            type='password'
            value={inputs.set_new_password_confirmation}
            onChange={(value) =>
              handleInputChange('set_new_password_confirmation', value)
            }
            size='large'
            className='!rounded-lg'
            prefix={<IconLock />}
          />
        </div>

        {turnstileEnabled && (
          <div className='flex justify-center'>
            <Turnstile
              sitekey={turnstileSiteKey}
              onVerify={(token) => {
                setTurnstileToken(token);
              }}
            />
          </div>
        )}
      </div>
    </Modal>
  );
};

export default ChangePasswordModal;
