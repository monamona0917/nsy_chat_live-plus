package dal

func GetChatRooms() ([]*ChatRoom, error) {
	innerChatRooms := make([]*ChatRoom, 0)
	err := db.Table(ChatRoom{}.TableName()).
		Order("display_name asc, id asc").
		Find(&innerChatRooms).Error
	return innerChatRooms, err
}

func GetChatRoomByChatRoomId(chatRoomId string) (*ChatRoom, error) {
	var room ChatRoom
	err := db.Table(ChatRoom{}.TableName()).
		Where("chat_room_id = ?", chatRoomId).
		Limit(1).
		Find(&room).Error
	if err != nil {
		return nil, err
	}
	if room.Id == 0 {
		return nil, nil
	}
	return &room, nil
}
