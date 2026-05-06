'use client'
import React from 'react';
import {Input} from '@/components/ui/input';
import {Button} from '@/components/ui/button';
import { Item } from '@/components/ui/item';
import {Field} from '@/components/ui/field';
import { useTranslation } from 'react-i18next';

export default function ProfilePage() {
  const [isEditing, setIsEditing] = React.useState(false);
  const [profileImage, setProfileImage] = React.useState('https://robohash.org/1.png?set=set1');
  const fileInputRef = React.useRef<HTMLInputElement>(null);
  const { t } = useTranslation();

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        setProfileImage(event.target?.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  return (
    <div className="container mx-auto p-4">
      <div className="flex mb-4 items-center justify-between gap-4">
        <h1 className="text-3xl font-bold">{t('Profile')}</h1>
        <Button 
          variant="default"
          onClick={() => setIsEditing(!isEditing)}
        >
          {isEditing ? "Enregistrer" : "Modifier"}
        </Button>
      </div>
      
      <div className="flex gap-8 items-start">
        <div className="flex flex-col gap-2 flex-1">
        <div>
          <label className="block text-sm font-medium ">{t('Email')}</label>
          <Item>
            <div className="flex">               
              <Input 
                className='bg-white'
                placeholder="Email"
                defaultValue="john.doe@example.com" 
                disabled={!isEditing}
              />
            </div>
          </Item>
        </div>

        <div>
          <label className="block text-sm font-medium ">{t('Name')}</label>
          <Item>
            <div className="flex gap-2 items-center">
              <Input 
                className='bg-white'
                placeholder="Nom" 
                defaultValue="Doe" 
                disabled={!isEditing}
              />
            </div>
          </Item>
        </div>
    
        <div>
          <label className="block text-sm font-medium ">{t('First name')}</label>
          <Item>
            <div className="flex gap-2 items-center">
              <Input 
                className='bg-white'
                placeholder="Prénom" 
                defaultValue="John" 
                disabled={!isEditing}
              />
            </div>
          </Item>
        </div>
        <div>
          <label className="block text-sm font-medium ">{t('Password')}</label>
          <Item>
            <div className="flex gap-2 items-center">
              <Input 
                className='bg-white'
                type="password"
                placeholder="Password" 
                disabled={!isEditing}
              />
            </div>
          </Item>
        </div>
        </div>

        {/* Photo de profil */}
        <div className="flex flex-col items-center gap-2 flex-shrink-0">
          <div 
            className={`relative w-40 h-40 rounded-lg overflow-hidden border-2 border-gray-300 ${isEditing ? 'cursor-pointer hover:border-blue-500' : ''}`}
            onClick={() => isEditing && fileInputRef.current?.click()}
          >
            <img 
              src={profileImage} 
              alt="Profil" 
              className="w-full h-full object-cover"
            />
            {isEditing && (
              <div className="absolute inset-0 bg-black bg-opacity-40 flex items-center justify-center">
                <span className="text-white text-sm font-medium">Modifier</span>
              </div>
            )}
          </div>
          <input 
            ref={fileInputRef}
            type="file" 
            accept="image/*"
            onChange={handleImageUpload}
            className="hidden"
          />
        </div>
      </div>
      
    </div>
  );
}
