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
import { Avatar, Card, Typography } from '@douyinfe/semi-ui';
import { Mail, Phone, User, Users } from 'lucide-react';

const PersonalInfoCard = ({ t, user }) => {
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

  return (
    <Card className='!rounded-2xl shadow-sm border-0'>
      <div className='flex items-center mb-4'>
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
                className='!mb-0 truncate'
                copyable={item.value && item.value !== t('未绑定') ? { content: item.value } : false}
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
