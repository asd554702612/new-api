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
import { Avatar, Button, Card, Tag, Typography } from '@douyinfe/semi-ui';
import {
  ExternalLink,
  Lock,
  Mail,
  Phone,
  RefreshCw,
  ShieldCheck,
  User,
  Users,
} from 'lucide-react';

const PersonalInfoCard = ({
  t,
  user,
  status,
  phoneVerificationEnabled,
  identitySyncLoading,
  identityVerificationLoading,
  onChangePhone,
  onChangePassword,
  onBindOIDC,
  onSyncIdentityStatus,
  onStartIdentityVerification,
}) => {
  const identityEnabled = Boolean(status?.casdoor_identity_enabled);
  const identityApiRequired = Boolean(status?.casdoor_identity_api_required);
  const oidcBound = Boolean(user?.oidc_id);
  const identityPassed = Boolean(
    user?.identity_verified &&
    user?.identity_age_checked &&
    user?.identity_over16,
  );

  const identityStatus = (() => {
    if (!oidcBound) {
      return {
        label: t('未绑定登录中心'),
        color: 'grey',
      };
    }
    if (identityPassed) {
      return {
        label: t('已实名认证'),
        color: 'green',
      };
    }
    return {
      label: t('未完成实名认证'),
      color: 'orange',
    };
  })();

  const identityActions = identityEnabled ? (
    <div className='mt-2 flex flex-wrap gap-2'>
      {!oidcBound ? (
        <Button
          type='primary'
          theme='outline'
          size='small'
          icon={<ExternalLink size={14} />}
          onClick={onBindOIDC}
          disabled={!status?.oidc_enabled}
        >
          {status?.oidc_enabled ? t('绑定登录中心') : t('登录中心未启用')}
        </Button>
      ) : (
        <>
          {!identityPassed && (
            <Button
              type='primary'
              theme='outline'
              size='small'
              icon={<ExternalLink size={14} />}
              loading={identityVerificationLoading}
              onClick={onStartIdentityVerification}
            >
              {t('去实名认证')}
            </Button>
          )}
          <Button
            theme='outline'
            size='small'
            icon={<RefreshCw size={14} />}
            loading={identitySyncLoading}
            onClick={onSyncIdentityStatus}
          >
            {t('刷新状态')}
          </Button>
        </>
      )}
    </div>
  ) : null;

  const identityApiNotice =
    identityEnabled && identityApiRequired ? (
      <div
        className={
          identityPassed
            ? 'mt-2 text-xs text-emerald-600'
            : 'mt-2 text-xs text-orange-600'
        }
      >
        {identityPassed
          ? t('已满足实名认证要求，API 调用可用')
          : t('管理员已要求完成实名认证后才能使用 API 调用')}
      </div>
    ) : null;

  const items = [
    {
      icon: <User size={18} />,
      label: t('用户名'),
      value: user?.username || t('未绑定'),
    },
    {
      icon: <Mail size={18} />,
      label: t('邮箱'),
      value: user?.email || t('未绑定'),
    },
    {
      icon: <Phone size={18} />,
      label: t('手机号'),
      value: user?.phone_number || t('未绑定'),
    },
    {
      icon: <Users size={18} />,
      label: t('用户分组'),
      value: user?.group || t('默认'),
    },
  ];

  if (identityEnabled) {
    items.push({
      icon: <ShieldCheck size={18} />,
      label: t('实名认证'),
      multiline: true,
      value: (
        <div>
          <Tag color={identityStatus.color}>{identityStatus.label}</Tag>
          {identityApiNotice}
          {identityActions}
        </div>
      ),
    });
  }

  return (
    <Card className='!rounded-2xl shadow-sm border-0'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between mb-4'>
        <div className='flex items-center'>
          <Avatar size='small' color='light-blue' className='mr-3 shadow-md'>
            <User size={16} />
          </Avatar>
          <div>
            <Typography.Text className='text-lg font-medium'>
              {t('个人信息')}
            </Typography.Text>
            <div className='text-xs text-gray-600'>
              {t('账号基础资料和联系方式')}
            </div>
          </div>
        </div>

        <div className='flex flex-wrap gap-2'>
          {phoneVerificationEnabled && (
            <Button
              type='primary'
              theme='outline'
              size='small'
              icon={<Phone size={14} />}
              onClick={onChangePhone}
            >
              {user?.phone_number ? t('修改手机号') : t('绑定手机号')}
            </Button>
          )}
          <Button
            type='primary'
            theme='outline'
            size='small'
            icon={<Lock size={14} />}
            onClick={onChangePassword}
          >
            {t('修改密码')}
          </Button>
        </div>
      </div>

      <div className='grid grid-cols-1 sm:grid-cols-2 gap-3'>
        {items.map((item) => (
          <div
            key={item.label}
            className='flex items-center gap-3 rounded-xl border border-gray-100 bg-gray-50/60 px-4 py-3'
          >
            <div className='flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-white text-slate-600 shadow-sm'>
              {item.icon}
            </div>
            <div className='min-w-0'>
              <Typography.Text type='tertiary' size='small'>
                {item.label}
              </Typography.Text>
              <Typography.Paragraph
                className={item.multiline ? '!mb-0' : '!mb-0 truncate'}
                copyable={
                  typeof item.value === 'string' &&
                  item.value &&
                  item.value !== t('未绑定')
                    ? { content: item.value }
                    : false
                }
              >
                {item.value}
              </Typography.Paragraph>
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
};

export default PersonalInfoCard;
